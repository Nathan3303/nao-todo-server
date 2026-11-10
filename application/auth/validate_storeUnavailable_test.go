package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/service"
	domaintypes "naotodoserver/domain/types"
)

// fakeValidateDomain 只实现 Validate 用到的三个领域方法，其余由内嵌接口占位
type fakeValidateDomain struct {
	service.IdentityDomain
	parseErr       error
	expired        bool
	session        *entities.UserSession
	sessionErr     error
	findSessionHit bool
}

func (f *fakeValidateDomain) ParseJWT(_ context.Context, _ string) (domaintypes.UserID, error) {
	if f.parseErr != nil {
		return 0, f.parseErr
	}
	return domaintypes.UserID(42), nil
}

func (f *fakeValidateDomain) IsJWTExpired(_ context.Context, _ string) bool {
	return f.expired
}

func (f *fakeValidateDomain) FindSessionByUserIdAndToken(
	_ context.Context, _ domaintypes.UserID, _ string,
) (*entities.UserSession, error) {
	f.findSessionHit = true
	return f.session, f.sessionErr
}

// TestValidate_StoreUnavailable_NotCredentialFailure 是 DEF-143 的应用层回归：
// 会话查询失败（DB/连接类）必须显式标记为 ErrStoreUnavailable，且**保留 err 原文**，
// 不得折叠成 ErrTokenExpired / ErrSessionNotFound（那会让中间件回 10041 ⇒ 客户端清认证）。
func TestValidate_StoreUnavailable_NotCredentialFailure(t *testing.T) {
	dbErr := errors.New(
		"sessionRepo.FindByUserIdAndToken: dial tcp 127.0.0.1:3306: connect: connection refused",
	)
	app := NewAuthApp(&fakeValidateDomain{sessionErr: dbErr}, nil, nil)

	_, err := app.Validate(context.Background(), "any-token")
	if err == nil {
		t.Fatal("会话查询失败时必须返回错误")
	}
	if !errors.Is(err, domerr.ErrStoreUnavailable) {
		t.Fatalf("会话查询失败应标记 ErrStoreUnavailable，实际: %v", err)
	}
	if errors.Is(err, domerr.ErrTokenExpired) || errors.Is(err, domerr.ErrSessionNotFound) {
		t.Fatalf("会话查询失败不得标记为凭证类错误，实际: %v", err)
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("必须保留 err 原文（AC-5），实际: %v", err)
	}
}

// TestValidate_CredentialFailures_Unchanged 覆盖 AC-4 的四类真失效语义不回归。
func TestValidate_CredentialFailures_Unchanged(t *testing.T) {
	cases := []struct {
		name   string
		domain *fakeValidateDomain
		want   error
	}{
		{
			name:   "JWT 签名或格式非法",
			domain: &fakeValidateDomain{parseErr: errors.New("signature is invalid")},
			want:   domerr.ErrTokenExpired,
		},
		{
			name:   "JWT 已过期",
			domain: &fakeValidateDomain{expired: true},
			want:   domerr.ErrTokenExpired,
		},
		{
			name:   "会话不存在（查不到）",
			domain: &fakeValidateDomain{sessionErr: domerr.ErrSessionNotFound},
			want:   domerr.ErrSessionNotFound,
		},
		{
			name:   "会话无效（过期行）",
			domain: &fakeValidateDomain{session: &entities.UserSession{UserId: 42}},
			want:   domerr.ErrTokenExpired,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			app := NewAuthApp(c.domain, nil, nil)
			_, err := app.Validate(context.Background(), "some-token")
			if !errors.Is(err, c.want) {
				t.Fatalf("期望 %v，实际: %v", c.want, err)
			}
			if errors.Is(err, domerr.ErrStoreUnavailable) {
				t.Fatalf("真凭证失效不得标记 ErrStoreUnavailable，实际: %v", err)
			}
		})
	}
}
