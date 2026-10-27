package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authApp "naotodoserver/application/auth"
	authDto "naotodoserver/application/auth/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// fakeSignUpAuth 只替换注册结果
type fakeSignUpAuth struct {
	authApp.AuthApp
	err error
}

func (f *fakeSignUpAuth) SignUp(_ context.Context, _ *authDto.SignUpInput) error {
	return f.err
}

// signUpBody 发一次注册请求
func signUpBody(t *testing.T, err error) (int, []byte) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	initTestLogger(t)
	engine := gin.New()
	ctrl := controllers.NewAuthController(&fakeSignUpAuth{err: err})
	engine.POST("/api/auth/signup", ctrl.UserSignUp)

	payload := []byte(`{"email":"victim@example.com","password":"12345678","nickname":"t570"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/signup", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

// TestSignUpWeakPasswordMapped 注册弱口令（SEC-05）⇒ 10001 + 统一可读文案，
// 且不回显命中了哪条规则/内部错误。
func TestSignUpWeakPasswordMapped(t *testing.T) {
	status, body := signUpBody(t, domerr.ErrPasswordTooWeak)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}

	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, body)
	}
	if res.Code != 10001 {
		t.Fatalf("code = %d, want 10001（注册参数错误码族）", res.Code)
	}
	if res.Message != "密码长度需 8–64 位且不能过于简单" {
		t.Fatalf("message = %q, want 统一可读文案", res.Message)
	}
	for _, leak := range []string{"弱口令", "名单", "本地部分", "全同", "纯数字", "bcrypt"} {
		if strings.Contains(res.Message, leak) {
			t.Fatalf("文案不得泄露规则细节（命中 %q）: %s", leak, res.Message)
		}
	}
}

// TestSignUpOtherFailureUnchanged 回归：非口令策略失败仍走既有 10002 / "注册失败"，不被误判成策略错误
func TestSignUpOtherFailureUnchanged(t *testing.T) {
	_, body := signUpBody(t, context.DeadlineExceeded)

	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, body)
	}
	if res.Code != 10002 || res.Message != "注册失败" {
		t.Fatalf("code/message = %d/%q, want 10002/注册失败", res.Code, res.Message)
	}
}
