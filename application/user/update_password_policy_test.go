package user

import (
	"context"
	"errors"
	"testing"

	"naotodoserver/application/user/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	domaintypes "naotodoserver/domain/types"
)

// fakePolicyUserRepo 只实现改密路径用到的两个方法，其余由嵌入接口兜底
type fakePolicyUserRepo struct {
	repositories.User
	user        *entities.User
	updateCalls int
	newPassword string
}

func (f *fakePolicyUserRepo) FindById(_ context.Context, _ domaintypes.UserID) (*entities.User, error) {
	if f.user == nil {
		return nil, errors.New("用户不存在")
	}
	return f.user, nil
}

func (f *fakePolicyUserRepo) UpdatePassword(
	_ context.Context,
	_ domaintypes.UserID,
	_ string,
	newPassword string,
) error {
	f.updateCalls++
	f.newPassword = newPassword
	return nil
}

// fakePolicySessionRepo 记录会话撤销调用
type fakePolicySessionRepo struct {
	repositories.UserSession
	revokeCalls int
}

func (f *fakePolicySessionRepo) DeleteByUserIdExceptToken(
	_ context.Context,
	_ domaintypes.UserID,
	_ string,
) (int64, error) {
	f.revokeCalls++
	return 0, nil
}

// TestUpdatePasswordRejectsWeakBeforeTouchingRepo 弱口令（SEC-05）：
// 在落库/撤销之前就被拒，既不改密也不误杀其它设备会话。
func TestUpdatePasswordRejectsWeakBeforeTouchingRepo(t *testing.T) {
	cases := map[string]string{
		"过短":       "abc1234",
		"过长":       "ab12cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"全同":       "aaaaaaaa",
		"纯数字短":     "12345678",
		"名单命中":     "password",
		"等于邮箱本地部分": "victim",
	}
	for name, newPassword := range cases {
		t.Run(name, func(t *testing.T) {
			userRepo := &fakePolicyUserRepo{user: &entities.User{Email: "victim@example.com"}}
			sessionRepo := &fakePolicySessionRepo{}
			app := NewUserApp(userRepo, sessionRepo, nil, nil)

			err := app.UpdatePassword(context.Background(), 42, dto.UpdatePasswordInput{
				OldPassword: "old-pass-1",
				NewPassword: newPassword,
				Token:       "current-token",
			})
			if !errors.Is(err, domerr.ErrPasswordTooWeak) {
				t.Fatalf("新口令 %q 应返回 ErrPasswordTooWeak，实际 %v", newPassword, err)
			}
			if userRepo.updateCalls != 0 {
				t.Fatalf("策略拒绝时不得落库，实际调用 %d 次", userRepo.updateCalls)
			}
			if sessionRepo.revokeCalls != 0 {
				t.Fatalf("策略拒绝时不得撤销会话，实际调用 %d 次", sessionRepo.revokeCalls)
			}
		})
	}
}

// TestUpdatePasswordAcceptsCompliantPolicy 合规新口令：正常落库并撤销其它设备
func TestUpdatePasswordAcceptsCompliantPolicy(t *testing.T) {
	userRepo := &fakePolicyUserRepo{user: &entities.User{Email: "victim@example.com"}}
	sessionRepo := &fakePolicySessionRepo{}
	app := NewUserApp(userRepo, sessionRepo, nil, nil)

	err := app.UpdatePassword(context.Background(), 42, dto.UpdatePasswordInput{
		OldPassword: "old-pass-1",
		NewPassword: "brand-new-pass-9",
		Token:       "current-token",
	})
	if err != nil {
		t.Fatalf("合规口令改密失败: %v", err)
	}
	if userRepo.updateCalls != 1 || userRepo.newPassword != "brand-new-pass-9" {
		t.Fatalf("落库调用 = %d / 新口令 = %q", userRepo.updateCalls, userRepo.newPassword)
	}
	if sessionRepo.revokeCalls != 1 {
		t.Fatalf("改密成功后应撤销其它设备，实际 %d 次", sessionRepo.revokeCalls)
	}
}
