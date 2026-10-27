package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"naotodoserver/application/auth/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/service"
	"naotodoserver/domain/identity/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/logging"

	"github.com/sirupsen/logrus"
)

// fakeSignInDomain 只实现登录节流与令牌签发，其余接口由内嵌接口占位
type fakeSignInDomain struct {
	service.IdentityDomain
	blocked      bool
	failures     int
	cleared      int
	blockedCalls int
}

func (f *fakeSignInDomain) IsLoginBlocked(_ context.Context, _ string) bool {
	f.blockedCalls++
	return f.blocked
}

func (f *fakeSignInDomain) RecordLoginFailure(_ context.Context, _ string) {
	f.failures++
}

func (f *fakeSignInDomain) ClearLoginFailures(_ context.Context, _ string) {
	f.cleared++
}

func (f *fakeSignInDomain) GenerateJWT(_ context.Context, _ *entities.User) (string, error) {
	return "jwt-token", nil
}

func (f *fakeSignInDomain) CreateSession(_ context.Context, _ domaintypes.UserID, _ string) error {
	return nil
}

// fakeSignInUserRepo 只实现登录所需的查库与口令比对
type fakeSignInUserRepo struct {
	repositories.User
	user      *entities.User
	findErr   error
	findCalls int
}

func (f *fakeSignInUserRepo) FindByEmail(_ context.Context, _ string) (*entities.User, error) {
	f.findCalls++
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.user, nil
}

func (f *fakeSignInUserRepo) PasswordCompare(password, hashed []byte) bool {
	return string(password) == string(hashed)
}

// captureAuthLogs 捕获应用层结构化日志
func captureAuthLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logging.Logger = logger
	t.Cleanup(func() {
		logging.Logger = previous
	})
	return buf
}

// findLogEntry 取指定 event 的日志字段
func findLogEntry(t *testing.T, buf *bytes.Buffer, event string) map[string]any {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry := map[string]any{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("日志不是结构化 JSON: %v（line=%s）", err, line)
		}
		if entry["event"] == event {
			return entry
		}
	}
	return nil
}

// TestSignInThrottledSkipsDBAndBcrypt 暂停期内：不查库、不做 bcrypt，直接返回密码等价错误；
// 日志只记哈希前缀，不含邮箱明文（SEC-04）。
func TestSignInThrottledSkipsDBAndBcrypt(t *testing.T) {
	logs := captureAuthLogs(t)
	domain := &fakeSignInDomain{blocked: true}
	repo := &fakeSignInUserRepo{user: &entities.User{Password: "right-pass"}}
	app := NewAuthApp(domain, repo, nil)

	_, err := app.SignIn(context.Background(), &dto.SignInInput{
		Email:    "Victim@Example.com",
		Password: "right-pass",
	})
	if !errors.Is(err, domerr.ErrLoginThrottled) {
		t.Fatalf("错误 = %v, want ErrLoginThrottled", err)
	}
	if repo.findCalls != 0 {
		t.Fatalf("暂停期内查库次数 = %d, want 0", repo.findCalls)
	}
	if domain.failures != 0 {
		t.Fatalf("暂停期内不应再累计失败: %d", domain.failures)
	}

	entry := findLogEntry(t, logs, "login.throttled")
	if entry == nil {
		t.Fatal("暂停必须记录 login.throttled 结构化日志")
	}
	if entry["key_tag"] != valueobjects.AccountKey("victim@example.com") {
		t.Fatalf("key_tag = %v, want %q", entry["key_tag"],
			valueobjects.AccountKey("victim@example.com"))
	}
	if strings.Contains(strings.ToLower(logs.String()), "victim@example.com") {
		t.Fatal("日志不得包含邮箱明文")
	}
}

// TestSignInPasswordMismatchRecordsFailure 口令错误 ⇒ 记录账号失败（对外仍是密码错误）
func TestSignInPasswordMismatchRecordsFailure(t *testing.T) {
	captureAuthLogs(t)
	domain := &fakeSignInDomain{}
	repo := &fakeSignInUserRepo{user: &entities.User{Password: "right-pass"}}
	app := NewAuthApp(domain, repo, nil)

	_, err := app.SignIn(context.Background(), &dto.SignInInput{
		Email:    "victim@example.com",
		Password: "wrong-pass",
	})
	if !errors.Is(err, domerr.ErrPasswordMismatch) {
		t.Fatalf("错误 = %v, want ErrPasswordMismatch", err)
	}
	if domain.failures != 1 {
		t.Fatalf("失败计数 = %d, want 1", domain.failures)
	}
	if domain.cleared != 0 {
		t.Fatalf("失败不应清除计数: %d", domain.cleared)
	}
}

// TestSignInSuccessClearsFailures 登录成功 ⇒ 清除该账号失败计数
func TestSignInSuccessClearsFailures(t *testing.T) {
	captureAuthLogs(t)
	domain := &fakeSignInDomain{}
	repo := &fakeSignInUserRepo{user: &entities.User{Password: "right-pass"}}
	app := NewAuthApp(domain, repo, nil)

	out, err := app.SignIn(context.Background(), &dto.SignInInput{
		Email:    "victim@example.com",
		Password: "right-pass",
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if out.Token != "jwt-token" {
		t.Fatalf("token = %q", out.Token)
	}
	if domain.cleared != 1 {
		t.Fatalf("成功登录应清除计数一次，实际 %d", domain.cleared)
	}
	if domain.failures != 0 {
		t.Fatalf("成功登录不应累计失败: %d", domain.failures)
	}
}

// TestSignInUnknownEmailDoesNotRecordFailure 不存在的邮箱不写入失败计数
// （避免任意邮箱消耗计数存储），对外仍与密码错误一致。
func TestSignInUnknownEmailDoesNotRecordFailure(t *testing.T) {
	captureAuthLogs(t)
	domain := &fakeSignInDomain{}
	repo := &fakeSignInUserRepo{findErr: domerr.ErrUserNotFound}
	app := NewAuthApp(domain, repo, nil)

	_, err := app.SignIn(context.Background(), &dto.SignInInput{
		Email:    "ghost@example.com",
		Password: "whatever",
	})
	if err == nil {
		t.Fatal("不存在的账号必须登录失败")
	}
	if domain.failures != 0 {
		t.Fatalf("不存在的邮箱不应写入失败计数: %d", domain.failures)
	}
}

// TestSignInThrottlePolicy 暴露给应用层的策略口径：5 次失败 / 5 分钟窗口 / 暂停 5 分钟
func TestSignInThrottlePolicy(t *testing.T) {
	if service.LoginFailureThreshold != 5 {
		t.Fatalf("失败阈值 = %d, want 5", service.LoginFailureThreshold)
	}
	if service.LoginFailureWindow != 5*time.Minute ||
		service.LoginBlockDuration != 5*time.Minute {
		t.Fatalf("窗口/暂停时长 = %v/%v, want 5m/5m",
			service.LoginFailureWindow, service.LoginBlockDuration)
	}
}
