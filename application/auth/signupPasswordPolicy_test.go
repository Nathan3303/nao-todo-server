package auth

import (
	"context"
	"errors"
	"testing"

	"naotodoserver/application/auth/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
)

// fakeSignUpUserRepo 只实现注册路径：邮箱不存在 + 记录创建调用
type fakeSignUpUserRepo struct {
	repositories.User
	createCalls int
	created     *valueobjects.CreateUser
}

func (f *fakeSignUpUserRepo) FindByEmail(_ context.Context, _ string) (*entities.User, error) {
	return nil, domerr.ErrUserNotFound
}

func (f *fakeSignUpUserRepo) CreateByVO(
	_ context.Context,
	vo *valueobjects.CreateUser,
) (*entities.User, error) {
	f.createCalls++
	f.created = vo
	return &entities.User{}, nil
}

// TestSignUpRejectsWeakPassword 注册弱口令（SEC-05）：领域值对象拒绝，不落库
func TestSignUpRejectsWeakPassword(t *testing.T) {
	cases := map[string]struct {
		email    string
		password string
	}{
		"过短":     {"t570@example.test", "abc1234"},
		"过长":     {"t570@example.test", "ab12cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"},
		"全同":     {"t570@example.test", "aaaaaaaa"},
		"纯数字短":   {"t570@example.test", "12345678"},
		"名单命中":   {"t570@example.test", "qwerty"},
		"等于邮箱本地": {"t570weak@example.test", "t570weak"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			userRepo := &fakeSignUpUserRepo{}
			app := NewAuthApp(nil, userRepo, nil)

			err := app.SignUp(context.Background(), &dto.SignUpInput{
				Email:    tc.email,
				Password: tc.password,
				Nickname: "t570",
			})
			if !errors.Is(err, domerr.ErrPasswordTooWeak) {
				t.Fatalf("弱口令 %q 应返回 ErrPasswordTooWeak，实际 %v", tc.password, err)
			}
			if userRepo.createCalls != 0 {
				t.Fatalf("策略拒绝时不得落库，实际 %d 次", userRepo.createCalls)
			}
		})
	}
}

// TestSignUpAcceptsCompliantPassword 合规口令：走到创建流程
func TestSignUpAcceptsCompliantPassword(t *testing.T) {
	userRepo := &fakeSignUpUserRepo{}
	app := NewAuthApp(nil, userRepo, nil)

	err := app.SignUp(context.Background(), &dto.SignUpInput{
		Email:    "t570-ok@example.test",
		Password: "correct-horse-9",
		Nickname: "t570",
	})
	if err != nil {
		t.Fatalf("合规口令注册失败: %v", err)
	}
	if userRepo.createCalls != 1 {
		t.Fatalf("合规口令应落库一次，实际 %d", userRepo.createCalls)
	}
}
