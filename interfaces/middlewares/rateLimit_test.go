package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authApp "naotodoserver/application/auth"
	iCtx "naotodoserver/infrastructure/context"

	"github.com/gin-gonic/gin"
)

// fakeUserRateLimitAuth 只实现 RateLimit（记录 Key/阈值，可按需返回超限错误）
type fakeUserRateLimitAuth struct {
	authApp.AuthApp
	err    error
	keys   []string
	limits []int64
}

func (f *fakeUserRateLimitAuth) RateLimit(_ context.Context, key string, limit int64) error {
	f.keys = append(f.keys, key)
	f.limits = append(f.limits, limit)
	return f.err
}

// newUserRateLimitEngine 模拟线上挂载顺序：先注入登录用户，再挂 UserRateLimiter
func newUserRateLimitEngine(auth *fakeUserRateLimitAuth, userId int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		if userId > 0 {
			ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), userId))
		}
		ctx.Next()
	})
	engine.Use(UserRateLimiter(auth, 120, "avatar"))
	engine.GET("/static/uploads/avatars/:filename", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})
	return engine
}

func doRateLimitRequest(t *testing.T, engine *gin.Engine, remoteAddr string) (int, int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/static/uploads/avatars/1.png", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var body struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code
}

// TestUserRateLimiterAllowsByUserKey 未超限放行，Key = user:<id>:avatar，阈值透传
func TestUserRateLimiterAllowsByUserKey(t *testing.T) {
	auth := &fakeUserRateLimitAuth{}
	status, _ := doRateLimitRequest(t, newUserRateLimitEngine(auth, 42), "10.0.0.1:1234")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if len(auth.keys) != 1 || auth.keys[0] != "user:42:avatar" {
		t.Fatalf("限流 Key = %v, want [user:42:avatar]", auth.keys)
	}
	if auth.limits[0] != 120 {
		t.Fatalf("阈值 = %d, want 120", auth.limits[0])
	}
}

// TestUserRateLimiterExceeded 超限 ⇒ 10051（与既有 RateLimiter 同错误码）
func TestUserRateLimiterExceeded(t *testing.T) {
	auth := &fakeUserRateLimitAuth{err: errors.New("超过配额")}
	status, code := doRateLimitRequest(t, newUserRateLimitEngine(auth, 42), "10.0.0.1:1234")
	if status != http.StatusOK || code != 10051 {
		t.Fatalf("超限应 10051，实际 status=%d code=%d", status, code)
	}
}

// TestUserRateLimiterRequiresUser 无登录用户 ⇒ 10041，且不消耗配额
func TestUserRateLimiterRequiresUser(t *testing.T) {
	auth := &fakeUserRateLimitAuth{}
	status, code := doRateLimitRequest(t, newUserRateLimitEngine(auth, 0), "10.0.0.1:1234")
	if status != http.StatusOK || code != 10041 {
		t.Fatalf("无用户应 10041，实际 status=%d code=%d", status, code)
	}
	if len(auth.keys) != 0 {
		t.Fatalf("无用户不应调用限流，实际 %v", auth.keys)
	}
}

// TestUserRateLimiterIgnoresIP 同一用户换 IP 仍落同一桶（不能靠换 IP 绕过）
func TestUserRateLimiterIgnoresIP(t *testing.T) {
	auth := &fakeUserRateLimitAuth{}
	engine := newUserRateLimitEngine(auth, 42)
	doRateLimitRequest(t, engine, "10.0.0.1:1234")
	doRateLimitRequest(t, engine, "192.168.1.9:9999")

	if len(auth.keys) != 2 ||
		auth.keys[0] != "user:42:avatar" || auth.keys[1] != "user:42:avatar" {
		t.Fatalf("不同 IP 应共享用户桶，实际 %v", auth.keys)
	}
}
