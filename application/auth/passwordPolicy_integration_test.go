//go:build integration

// SEC-05 / T570 端到端回归：真实 MySQL + Redis，生产同构路由。
// 注册与改密共用同一口令策略；既有弱口令用户登录不受影响（不强制改密）。
package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authApp "naotodoserver/application/auth"
	userApp "naotodoserver/application/user"
	"naotodoserver/domain/identity/service"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"
	"naotodoserver/interfaces/routers"

	"github.com/gin-gonic/gin"
)

// policyMessage 口令策略不通过时的统一对外文案
const policyMessage = "密码长度需 8–64 位且不能过于简单"

// newPasswordPolicyEngine 生产同构装配：auth 路由（注册/登录）+ 改密路由
func newPasswordPolicyEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	userRepo := identity.NewUserRepo(itDB, cache.NewCache(itRds))
	sessionRepo := identity.NewSessionRepo(itDB, cache.NewCache(itRds))
	localFallback := identity.NewLocalFallback()
	domain := service.NewIdentityDomain(
		identity.NewJWTRepo(),
		sessionRepo,
		identity.NewRateLimitRepo(itRds, localFallback),
		identity.NewLoginThrottleRepo(itRds, localFallback),
	)
	auth := authApp.NewAuthApp(domain, userRepo, sessionRepo)
	engine := gin.New()
	api := engine.Group("/api")
	routers.UseAuthRouter(api, controllers.NewAuthController(auth), auth)
	api.PUT(
		"/user/password",
		middlewares.JWTValidator(auth),
		controllers.NewUserController(userApp.NewUserApp(userRepo, sessionRepo, nil, nil)).
			UpdateUserPassword,
	)
	return engine
}

// policyRes 统一响应解析
type policyRes struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// postJSON 发送 JSON 请求（可带 Bearer 令牌）
func postJSON(
	t *testing.T, engine *gin.Engine, method, path, ip, token string, payload any,
) policyRes {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("序列化请求失败: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.RemoteAddr = ip
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	res := policyRes{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	return res
}

// signupPolicy 发一次注册
func signupPolicy(t *testing.T, engine *gin.Engine, ip, email, password string) policyRes {
	t.Helper()
	return postJSON(t, engine, http.MethodPost, "/api/auth/signup", ip, "", map[string]string{
		"email": email, "password": password, "nickname": "t570",
	})
}

// signinPolicy 发一次登录
func signinPolicy(t *testing.T, engine *gin.Engine, ip, email, password string) policyRes {
	t.Helper()
	return postJSON(t, engine, http.MethodPost, "/api/auth/signin", ip, "", map[string]string{
		"email": email, "password": password,
	})
}

// changePasswordPolicy 发一次改密（需登录态）
func changePasswordPolicy(
	t *testing.T, engine *gin.Engine, ip, token, oldPassword, newPassword string,
) policyRes {
	t.Helper()
	return postJSON(t, engine, http.MethodPut, "/api/user/password", ip, token, map[string]string{
		"oldPassword": oldPassword, "newPassword": newPassword,
	})
}

// TestSignUpPasswordPolicyEndToEnd 注册：弱口令矩阵被拒（10001 + 统一文案），合规口令通过
func TestSignUpPasswordPolicyEndToEnd(t *testing.T) {
	captureLogs(t)
	engine := newPasswordPolicyEngine()
	ip := "203.0.113.70:40100"
	local := fmt.Sprintf("t570pw-%d", time.Now().UnixNano())

	weak := map[string]string{
		"过短":     "abc1234",
		"过长":     "ab12" + strings.Repeat("c", 64),
		"全同":     "aaaaaaaa",
		"纯数字短":   "12345678",
		"名单命中":   "qwerty",
		"等于邮箱本地": local,
	}
	for name, password := range weak {
		res := signupPolicy(t, engine, ip, local+"@example.test", password)
		if res.Code != 10001 {
			t.Fatalf("%s: code = %d, want 10001（body=%+v）", name, res.Code, res)
		}
		if res.Message != policyMessage {
			t.Fatalf("%s: message = %q, want %q", name, res.Message, policyMessage)
		}
	}

	okEmail := uniqueEmail("t570pw-ok")
	if res := signupPolicy(t, engine, ip, okEmail, "correct-horse-9"); res.Code != 10000 {
		t.Fatalf("合规口令注册 code = %d, want 10000（body=%+v）", res.Code, res)
	}
	if res := signinPolicy(t, engine, ip+"-2", okEmail, "correct-horse-9"); res.Code != 10010 {
		t.Fatalf("合规口令注册后登录 code = %d, want 10010", res.Code)
	}
}

// TestChangePasswordPolicyEndToEnd 改密：弱口令被拒（10073 + 统一文案）；合规口令生效；
// 既有弱口令用户仍可登录（兼容：旧口令不强制改）
func TestChangePasswordPolicyEndToEnd(t *testing.T) {
	captureLogs(t)
	engine := newPasswordPolicyEngine()
	ip := "203.0.113.71:40101"
	email := uniqueEmail("t570chg")
	// 策略上线前注册的旧用户：口令 "1"（按新策略不合规，但登录不应被拦）
	seedUser(t, email, "1")

	res := signinPolicy(t, engine, ip, email, "1")
	if res.Code != 10010 {
		t.Fatalf("旧弱口令用户登录 code = %d, want 10010（兼容：登录路径不校验策略）", res.Code)
	}
	var out struct {
		Token string `json:"jwt"`
	}
	if err := json.Unmarshal(res.Data, &out); err != nil || out.Token == "" {
		t.Fatalf("登录未返回 token: %v（data=%s）", err, res.Data)
	}
	token := out.Token

	weak := map[string]string{
		"过短":   "abc1234",
		"过长":   "ab12" + strings.Repeat("c", 64),
		"全同":   "aaaaaaaa",
		"纯数字短": "12345678",
		"名单命中": "password",
	}
	for name, newPassword := range weak {
		got := changePasswordPolicy(t, engine, ip, token, "1", newPassword)
		if got.Code != 10073 {
			t.Fatalf("%s: 改密 code = %d, want 10073（body=%+v）", name, got.Code, got)
		}
		if got.Message != policyMessage {
			t.Fatalf("%s: message = %q, want %q", name, got.Message, policyMessage)
		}
	}

	if got := changePasswordPolicy(t, engine, ip, token, "1", "brand-new-pass-9"); got.Code != 10070 {
		t.Fatalf("合规口令改密 code = %d, want 10070（body=%+v）", got.Code, got)
	}
	if got := signinPolicy(t, engine, ip, email, "1"); got.Code != 10012 {
		t.Fatalf("改密后旧口令登录 code = %d, want 10012", got.Code)
	}
	if got := signinPolicy(t, engine, ip, email, "brand-new-pass-9"); got.Code != 10010 {
		t.Fatalf("改密后新口令登录 code = %d, want 10010", got.Code)
	}
}
