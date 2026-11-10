// #88 / T656：/sync 限流双桶（IP 粗桶 → JWT → user pull/push 桶）回归。
package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	authApp "naotodoserver/application/auth"
	"naotodoserver/conf"
	domerr "naotodoserver/domain/errors"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// fakeSyncRouterAuth 提供 Validate / RateLimit：
//   - Validate 记录调用次数（C1：未鉴权洪水不得越过 IP 桶打到验签/查库）
//   - RateLimit 按 key 做固定窗口计数（与线上同一口径：第 limit+1 次拒绝）
type fakeSyncRouterAuth struct {
	authApp.AuthApp

	mu            sync.Mutex
	userId        int64
	validateCalls int
	counts        map[string]int64
	limits        map[string]int64
}

func (f *fakeSyncRouterAuth) Validate(_ context.Context, _ string) (
	domaintypes.UserID, error,
) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.validateCalls++
	if f.userId <= 0 {
		// 夹具对齐 DEF-143：凭证类必须用显式哨兵（断言值仍为 10041）
		return 0, domerr.ErrTokenExpired
	}
	return domaintypes.UserID(f.userId), nil
}

func (f *fakeSyncRouterAuth) RateLimit(_ context.Context, key string, limit int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.counts == nil {
		f.counts = map[string]int64{}
	}
	if f.limits == nil {
		f.limits = map[string]int64{}
	}
	f.limits[key] = limit
	f.counts[key]++
	if f.counts[key] > limit {
		return errors.New("超过配额")
	}
	return nil
}

func (f *fakeSyncRouterAuth) validateCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.validateCalls
}

func (f *fakeSyncRouterAuth) limitOf(key string) int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.limits[key]
}

// syncRouterFixture 构造生产默认配额（120 / 12 / 24）
func syncRouterFixture(t *testing.T) {
	t.Helper()
	old := conf.Conf
	conf.Conf = &conf.Config{Sync: conf.Sync{RateLimit: conf.SyncRateLimit{
		IPPerMin:   120,
		PullPerMin: 12,
		PushPerMin: 24,
	}}}
	t.Cleanup(func() { conf.Conf = old })
}

func newSyncRouterEngine(auth *fakeSyncRouterAuth) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewSyncController(nil, nil, nil, nil, nil, nil)
	UseSyncRouter(engine.Group("/api"), ctrl, auth)
	return engine
}

func doSyncRouterRequest(t *testing.T, engine *gin.Engine, path string) int {
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

// TestSyncRoutesUnchanged C4：契约路径不变，仍是 POST /api/sync/pull 与 /api/sync/push
func TestSyncRoutesUnchanged(t *testing.T) {
	syncRouterFixture(t)
	engine := newSyncRouterEngine(&fakeSyncRouterAuth{})

	routes := make(map[string]bool, len(engine.Routes()))
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, want := range []string{"POST /api/sync/pull", "POST /api/sync/push"} {
		if !routes[want] {
			t.Fatalf("缺少同步路由 %q（契约变更）", want)
		}
	}
	if len(routes) != 2 {
		t.Fatalf("同步路由数 = %d, want 2（不得新增路径）", len(routes))
	}
}

// TestSyncUnauthenticatedFloodStoppedBeforeValidate C1：
// 同一 IP 未鉴权洪水 130 次 ⇒ 仅前 120 次进入 Validate（验签+查库），其余被 IP 粗桶拦下。
func TestSyncUnauthenticatedFloodStoppedBeforeValidate(t *testing.T) {
	syncRouterFixture(t)
	auth := &fakeSyncRouterAuth{} // 凭证非法
	engine := newSyncRouterEngine(auth)

	blocked := 0
	for i := 0; i < 130; i++ {
		if code := doSyncRouterRequest(t, engine, "/api/sync/pull"); code == 10051 {
			blocked++
		}
	}

	if got := auth.validateCount(); got != 120 {
		t.Fatalf("Validate 调用次数 = %d, want 120（超出 IP 配额的洪水不得触 DB）", got)
	}
	if blocked != 10 {
		t.Fatalf("IP 桶拦截数 = %d, want 10", blocked)
	}
	if got := auth.limitOf("10.0.0.1:sync-ip"); got != 120 {
		t.Fatalf("IP 粗桶配额 = %d, want 120", got)
	}
}

// TestSyncRateLimitThreeStateCodes C3：
// 有效未超限=放行；有效超限=10051；凭证无效（即使桶已打满）=10041。
func TestSyncRateLimitThreeStateCodes(t *testing.T) {
	syncRouterFixture(t)

	// 有效未超限：越过 user 桶进入控制器；控制器以 nil 依赖会 panic，
	// 故用「最后一次 UserRateLimiter 已放行」的间接证据（Validate 成功 + 无 10051/10041）。
	// 这里只验证拒绝两态，放行路径由 middlewares 单测覆盖，避免触发真实 handler。
	auth := &fakeSyncRouterAuth{userId: 7}
	engine := newSyncRouterEngine(auth)
	auth.mu.Lock()
	auth.counts = map[string]int64{"user:7:sync-pull": 12}
	auth.mu.Unlock()

	if code := doSyncRouterRequest(t, engine, "/api/sync/pull"); code != 10051 {
		t.Fatalf("有效凭证超限应 10051，实际 %d", code)
	}
	if got := auth.limitOf("user:7:sync-pull"); got != 12 {
		t.Fatalf("user pull 桶配额 = %d, want 12", got)
	}

	invalidAuth := &fakeSyncRouterAuth{} // 凭证无效
	invalidEngine := newSyncRouterEngine(invalidAuth)
	invalidAuth.mu.Lock()
	invalidAuth.counts = map[string]int64{"user:7:sync-pull": 12}
	invalidAuth.mu.Unlock()

	if code := doSyncRouterRequest(t, invalidEngine, "/api/sync/pull"); code != 10041 {
		t.Fatalf("凭证无效应 10041（不得因超限变 10051），实际 %d", code)
	}
}
