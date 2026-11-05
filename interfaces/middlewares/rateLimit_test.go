package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authApp "naotodoserver/application/auth"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/logging"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

// fakeSyncRateLimitAuth 同时提供 Validate（C1 计数用）与 RateLimit（分桶用）
type fakeSyncRateLimitAuth struct {
	authApp.AuthApp
	userId        int64
	validateCalls int
	denyKeys      map[string]bool
	keys          []string
	limits        []int64
}

func (f *fakeSyncRateLimitAuth) Validate(_ context.Context, _ string) (
	domaintypes.UserID, error,
) {
	f.validateCalls++
	if f.userId <= 0 {
		return 0, errors.New("invalid token")
	}
	return domaintypes.UserID(f.userId), nil
}

func (f *fakeSyncRateLimitAuth) RateLimit(_ context.Context, key string, limit int64) error {
	f.keys = append(f.keys, key)
	f.limits = append(f.limits, limit)
	if f.denyKeys[key] {
		return errors.New("超过配额")
	}
	return nil
}

// newSyncBucketEngine 复刻 /sync 鉴权后的分桶挂载：pull / push 各自 UserRateLimiter
func newSyncBucketEngine(
	auth *fakeSyncRateLimitAuth,
	pullLimit int64,
	pushLimit int64,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(
			iCtx.SetUserId(ctx.Request.Context(), auth.userId),
		)
		ctx.Next()
	})
	ok := func(ctx *gin.Context) { ctx.String(http.StatusOK, "ok") }
	engine.POST("/api/sync/pull", UserRateLimiter(auth, pullLimit, "sync-pull"), ok)
	engine.POST("/api/sync/push", UserRateLimiter(auth, pushLimit, "sync-push"), ok)
	return engine
}

func doSyncBucketRequest(t *testing.T, engine *gin.Engine, path string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var body struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return body.Code
}

// TestSyncBucketsIndependent #88 / C2：打满 pull 不影响 push，且两桶配额独立透传（12 / 24）
func TestSyncBucketsIndependent(t *testing.T) {
	auth := &fakeSyncRateLimitAuth{
		userId:   7,
		denyKeys: map[string]bool{"user:7:sync-pull": true},
	}
	engine := newSyncBucketEngine(auth, 12, 24)

	if code := doSyncBucketRequest(t, engine, "/api/sync/pull"); code != 10051 {
		t.Fatalf("pull 超限应 10051，实际 %d", code)
	}
	if code := doSyncBucketRequest(t, engine, "/api/sync/push"); code != 0 {
		t.Fatalf("push 不应受 pull 超限影响，实际业务码 %d", code)
	}

	limits := map[string]int64{}
	for i, key := range auth.keys {
		limits[key] = auth.limits[i]
	}
	if limits["user:7:sync-pull"] != 12 {
		t.Fatalf("pull 配额 = %d, want 12", limits["user:7:sync-pull"])
	}
	if limits["user:7:sync-push"] != 24 {
		t.Fatalf("push 配额 = %d, want 24", limits["user:7:sync-push"])
	}
}

// TestSyncUserBucketKeyShape #88 / C6：key = user:<id>:<tag>，不含 token / 邮箱
func TestSyncUserBucketKeyShape(t *testing.T) {
	auth := &fakeSyncRateLimitAuth{userId: 42}
	engine := newSyncBucketEngine(auth, 12, 24)

	if code := doSyncBucketRequest(t, engine, "/api/sync/pull"); code != 0 {
		t.Fatalf("未超限应放行，实际业务码 %d", code)
	}
	if len(auth.keys) != 1 || auth.keys[0] != "user:42:sync-pull" {
		t.Fatalf("限流 Key = %v, want [user:42:sync-pull]", auth.keys)
	}
	if strings.Contains(auth.keys[0], "@") || strings.Contains(auth.keys[0], "token") {
		t.Fatalf("限流 Key 不得包含 token / 邮箱: %q", auth.keys[0])
	}
}

// captureRateLimitLogs 捕获结构化日志（只在本文件用）
func captureRateLimitLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	logging.Logger = logger
	t.Cleanup(func() { logging.Logger = previous })
	return buf
}

// TestRateLimitBlockedLogRedacted #88 / C7：拒绝时落 ratelimit.blocked 计数，
// 只含 key_tag 摘要与桶名，不含明文 userId / token。
func TestRateLimitBlockedLogRedacted(t *testing.T) {
	logs := captureRateLimitLogs(t)
	auth := &fakeSyncRateLimitAuth{
		userId:   42,
		denyKeys: map[string]bool{"user:42:sync-push": true},
	}
	engine := newSyncBucketEngine(auth, 12, 24)

	if code := doSyncBucketRequest(t, engine, "/api/sync/push"); code != 10051 {
		t.Fatalf("push 超限应 10051，实际 %d", code)
	}

	raw := logs.String()
	if strings.Contains(raw, "user:42") {
		t.Fatalf("限流日志不得包含明文 key / userId: %s", raw)
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &entry); err != nil {
		t.Fatalf("限流日志不是结构化 JSON: %v（%s）", err, raw)
	}
	if entry["event"] != "ratelimit.blocked" {
		t.Fatalf("event = %v, want ratelimit.blocked", entry["event"])
	}
	if entry["bucket"] != "sync-push" {
		t.Fatalf("bucket = %v, want sync-push", entry["bucket"])
	}
	if tag, ok := entry["key_tag"].(string); !ok || tag == "" {
		t.Fatalf("key_tag 缺失: %v", entry)
	}
}
