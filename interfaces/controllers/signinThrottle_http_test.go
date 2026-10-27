package controllers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authApp "naotodoserver/application/auth"
	authDto "naotodoserver/application/auth/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// initTestLogger 控制器失败分支会写日志：注入捕获型日志器，避免全局 Logger 为 nil
func initTestLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	logger := logrus.New()
	logger.SetOutput(buf)
	logging.Logger = logger
	t.Cleanup(func() { logging.Logger = previous })
	return buf
}

// fakeSignInAuth 只替换 SignIn 结果，其余接口由内嵌接口占位
type fakeSignInAuth struct {
	authApp.AuthApp
	err error
}

func (f *fakeSignInAuth) SignIn(
	_ context.Context,
	_ *authDto.SignInInput,
) (*authDto.SignInOutput, error) {
	return nil, f.err
}

// signInBody 发一次登录请求，返回状态码与响应原始字节
func signInBody(t *testing.T, err error) (int, []byte) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewAuthController(&fakeSignInAuth{err: err})
	engine.POST("/api/auth/signin", ctrl.UserSignIn)

	payload := []byte(`{"email":"victim@example.com","password":"wrong-pass"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

// TestSignInThrottledResponseEqualsPasswordMismatch 账号暂停（SEC-04）与密码错误的响应
// 必须逐字节一致：状态码、code、message、data 均无差异（防账号枚举）。
func TestSignInThrottledResponseEqualsPasswordMismatch(t *testing.T) {
	logs := initTestLogger(t)
	// 先跑暂停分支：该路径不得落邮箱明文（日志只由应用层记哈希前缀）
	throttledStatus, throttledBody := signInBody(t, domerr.ErrLoginThrottled)
	if strings.Contains(logs.String(), "victim@example.com") {
		t.Fatalf("暂停响应不得在日志中落邮箱明文: %s", logs.String())
	}

	mismatchStatus, mismatchBody := signInBody(t, domerr.ErrPasswordMismatch)
	_, notFoundBody := signInBody(t, domerr.ErrUserNotFound)

	if mismatchStatus != http.StatusOK || throttledStatus != mismatchStatus {
		t.Fatalf("状态码不一致: mismatch=%d throttled=%d", mismatchStatus, throttledStatus)
	}
	if !bytes.Equal(throttledBody, mismatchBody) {
		t.Fatalf("暂停响应必须与密码错误逐字节一致:\n mismatch=%s\n throttled=%s",
			mismatchBody, throttledBody)
	}
	if !bytes.Equal(notFoundBody, mismatchBody) {
		t.Fatalf("用户不存在响应必须与密码错误逐字节一致:\n notfound=%s\n mismatch=%s",
			notFoundBody, mismatchBody)
	}
	if !bytes.Contains(mismatchBody, []byte(`"code":10012`)) {
		t.Fatalf("密码错误响应码应为 10012: %s", mismatchBody)
	}
	if bytes.Contains(throttledBody, []byte("jwt")) {
		t.Fatalf("暂停响应不得携带凭证: %s", throttledBody)
	}
}
