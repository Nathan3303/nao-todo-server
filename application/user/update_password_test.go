package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"naotodoserver/application/user/dto"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	domaintypes "naotodoserver/domain/types"
)

// fakeUserRepo 只实现改密路径；其余方法由嵌入接口兜底（本测试不会调用）。
type fakeUserRepo struct {
	repositories.User
	email       string
	password    string
	updateErr   error
	updateCalls int
}

// FindById 口令策略（SEC-05）需要用户邮箱做"不等于邮箱本地部分"判定
func (f *fakeUserRepo) FindById(_ context.Context, _ domaintypes.UserID) (*entities.User, error) {
	return &entities.User{Email: f.email}, nil
}

func (f *fakeUserRepo) UpdatePassword(
	_ context.Context,
	_ domaintypes.UserID,
	oldPassword, newPassword string,
) error {
	f.updateCalls++
	if f.updateErr != nil {
		return f.updateErr
	}
	if oldPassword != f.password {
		return errors.New("旧密码错误")
	}
	f.password = newPassword
	return nil
}

// fakeSessionRepo 用内存集合模拟会话表：撤销即从集合移除。
type fakeSessionRepo struct {
	repositories.UserSession
	sessions    map[string]struct{}
	revokeErr   error
	revokeCalls int
	lastKeep    string
}

func newFakeSessionRepo(tokens ...string) *fakeSessionRepo {
	repo := &fakeSessionRepo{sessions: make(map[string]struct{}, len(tokens))}
	for _, token := range tokens {
		repo.sessions[token] = struct{}{}
	}
	return repo
}

func (f *fakeSessionRepo) DeleteByUserIdExceptToken(
	_ context.Context,
	_ domaintypes.UserID,
	keepToken string,
) (int64, error) {
	f.revokeCalls++
	f.lastKeep = keepToken
	if f.revokeErr != nil {
		return 0, f.revokeErr
	}
	var revoked int64
	for token := range f.sessions {
		if token != keepToken {
			delete(f.sessions, token)
			revoked++
		}
	}
	return revoked, nil
}

// has 判断会话是否仍存在
func (f *fakeSessionRepo) has(token string) bool {
	_, ok := f.sessions[token]
	return ok
}

// newTestUserApp 构造仅用于改密路径的应用实例（task/avatar 依赖不参与）
func newTestUserApp(userRepo *fakeUserRepo, sessionRepo *fakeSessionRepo) UserApp {
	return NewUserApp(userRepo, sessionRepo, nil, nil)
}

// TestUpdatePasswordRevokesOtherSessions 改密成功：撤销其它设备、保留当前设备
func TestUpdatePasswordRevokesOtherSessions(t *testing.T) {
	userRepo := &fakeUserRepo{password: "old-pass-1"}
	sessionRepo := newFakeSessionRepo("current-token", "other-token-1", "other-token-2")
	app := newTestUserApp(userRepo, sessionRepo)

	err := app.UpdatePassword(context.Background(), 42, dto.UpdatePasswordInput{
		OldPassword: "old-pass-1",
		NewPassword: "new-pass-1",
		Token:       "current-token",
	})
	if err != nil {
		t.Fatalf("改密失败: %v", err)
	}
	if !sessionRepo.has("current-token") {
		t.Fatal("当前设备会话必须保留")
	}
	if sessionRepo.has("other-token-1") || sessionRepo.has("other-token-2") {
		t.Fatalf("其它设备会话必须撤销，实际残留: %v", sessionRepo.sessions)
	}
	if sessionRepo.revokeCalls != 1 || sessionRepo.lastKeep != "current-token" {
		t.Fatalf("撤销调用 = %d 次 / keep=%q, want 1 次 / current-token",
			sessionRepo.revokeCalls, sessionRepo.lastKeep)
	}
	if userRepo.password != "new-pass-1" {
		t.Fatalf("密码未更新，实际 %q", userRepo.password)
	}
}

// TestUpdatePasswordWrongOldPasswordKeepsSessions 旧密码错 ⇒ 不改密、不撤销任何会话
func TestUpdatePasswordWrongOldPasswordKeepsSessions(t *testing.T) {
	userRepo := &fakeUserRepo{password: "old-pass-1"}
	sessionRepo := newFakeSessionRepo("current-token", "other-token")
	app := newTestUserApp(userRepo, sessionRepo)

	err := app.UpdatePassword(context.Background(), 42, dto.UpdatePasswordInput{
		OldPassword: "wrong-pass",
		NewPassword: "new-pass-1",
		Token:       "current-token",
	})
	if err == nil {
		t.Fatal("旧密码错误必须返回失败")
	}
	if sessionRepo.revokeCalls != 0 {
		t.Fatalf("改密失败不得撤销会话，实际调用 %d 次", sessionRepo.revokeCalls)
	}
	if !sessionRepo.has("other-token") {
		t.Fatal("改密失败后其它设备会话必须仍有效")
	}
}

// TestUpdatePasswordRevokeFailureSurfaces 密码已更新但撤销失败 ⇒ 如实返回错误（不静默）
func TestUpdatePasswordRevokeFailureSurfaces(t *testing.T) {
	userRepo := &fakeUserRepo{password: "old-pass-1"}
	sessionRepo := newFakeSessionRepo("current-token", "other-token")
	sessionRepo.revokeErr = errors.New("db down")
	app := newTestUserApp(userRepo, sessionRepo)

	err := app.UpdatePassword(context.Background(), 42, dto.UpdatePasswordInput{
		OldPassword: "old-pass-1",
		NewPassword: "new-pass-1",
		Token:       "current-token",
	})
	if err == nil || !strings.Contains(err.Error(), "RevokeOtherSessions") {
		t.Fatalf("撤销失败必须上抛，实际: %v", err)
	}
}

// TestUpdatePasswordRetryDoesNotChangeSemantics 幂等重试：同一请求再次提交
// ⇒ 旧密码已失效而失败，不再产生撤销、当前设备仍在线
func TestUpdatePasswordRetryDoesNotChangeSemantics(t *testing.T) {
	userRepo := &fakeUserRepo{password: "old-pass-1"}
	sessionRepo := newFakeSessionRepo("current-token", "other-token")
	app := newTestUserApp(userRepo, sessionRepo)
	req := dto.UpdatePasswordInput{
		OldPassword: "old-pass-1",
		NewPassword: "new-pass-1",
		Token:       "current-token",
	}

	if err := app.UpdatePassword(context.Background(), 42, req); err != nil {
		t.Fatalf("首次改密失败: %v", err)
	}
	if err := app.UpdatePassword(context.Background(), 42, req); err == nil {
		t.Fatal("重试应因旧密码已失效而失败")
	}
	if sessionRepo.revokeCalls != 1 {
		t.Fatalf("重试不得再次撤销，实际调用 %d 次", sessionRepo.revokeCalls)
	}
	if !sessionRepo.has("current-token") || sessionRepo.has("other-token") {
		t.Fatalf("重试后会话状态被改变: %v", sessionRepo.sessions)
	}
}

// TestUpdatePasswordWithoutTokenRevokesAll 无当前令牌（理论上鉴权路由不会发生）
// ⇒ 退化为全部撤销并要求重新登录，保证改密后旧会话必然失效
func TestUpdatePasswordWithoutTokenRevokesAll(t *testing.T) {
	userRepo := &fakeUserRepo{password: "old-pass-1"}
	sessionRepo := newFakeSessionRepo("token-a", "token-b")
	app := newTestUserApp(userRepo, sessionRepo)

	err := app.UpdatePassword(context.Background(), 42, dto.UpdatePasswordInput{
		OldPassword: "old-pass-1",
		NewPassword: "new-pass-1",
	})
	if err != nil {
		t.Fatalf("改密失败: %v", err)
	}
	if len(sessionRepo.sessions) != 0 {
		t.Fatalf("无当前令牌时应全部撤销，实际残留: %v", sessionRepo.sessions)
	}
}
