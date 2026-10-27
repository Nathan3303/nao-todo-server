package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authApp "naotodoserver/application/auth"
	domaintypes "naotodoserver/domain/types"

	"github.com/gin-gonic/gin"
)

// fakeAuthApp 只实现 Validate：valid 集合内的 token 视为有效
type fakeAuthApp struct {
	authApp.AuthApp
	valid map[string]bool
}

func (f *fakeAuthApp) Validate(_ context.Context, token string) (domaintypes.UserID, error) {
	if f.valid[token] {
		return domaintypes.UserID(42), nil
	}
	return 0, errors.New("invalid token")
}

// newJWTTestEngine 注册与线上同形的路由：/api 组 + 头像路由 + 前缀边界路由
func newJWTTestEngine(valid ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	app := &fakeAuthApp{valid: make(map[string]bool, len(valid))}
	for _, token := range valid {
		app.valid[token] = true
	}
	ok := func(ctx *gin.Context) { ctx.String(http.StatusOK, "ok") }

	engine := gin.New()
	api := engine.Group("/api", JWTValidator(app))
	api.GET("/tasks", ok)         // 普通 API（不允许 query token）
	api.GET("/sse/reminders", ok) // SSE（允许 query token）
	api.GET("/ssevil", ok)        // 前缀边界：不得被 /api/sse 前缀放过
	engine.GET("/static/uploads/avatars/:filename", JWTValidator(app), ok)
	// 前缀边界：不得被 /static/uploads/avatars/ 前缀放过
	engine.GET("/static/uploads/avatars-export", JWTValidator(app), ok)
	return engine
}

type jwtTestResponse struct {
	Code int `json:"code"`
}

// doJWTRequest 发请求并返回 HTTP 状态 + 业务 code（成功处理器返回纯文本 ⇒ code=0）
func doJWTRequest(t *testing.T, engine *gin.Engine, target, authHeader string) (int, int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var body jwtTestResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code
}

// TestJWTValidatorQueryTokenAllowlisted 头像 / SSE 仍可用 ?token=（兼容性自证）
func TestJWTValidatorQueryTokenAllowlisted(t *testing.T) {
	engine := newJWTTestEngine("query-token")
	cases := map[string]string{
		"头像":  "/static/uploads/avatars/1.png?token=query-token",
		"SSE": "/api/sse/reminders?token=query-token",
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			status, code := doJWTRequest(t, engine, target, "")
			if status != http.StatusOK || code != 0 {
				t.Fatalf("%s 应放行，实际 status=%d code=%d", name, status, code)
			}
		})
	}
}

// TestJWTValidatorPlainAPIIgnoresQueryToken 普通 API 仅 query ⇒ 10041
func TestJWTValidatorPlainAPIIgnoresQueryToken(t *testing.T) {
	engine := newJWTTestEngine("query-token")
	status, code := doJWTRequest(t, engine, "/api/tasks?token=query-token", "")
	if status != http.StatusOK || code != 10041 {
		t.Fatalf("仅 query 应被拒为 10041，实际 status=%d code=%d", status, code)
	}
}

// TestJWTValidatorHeaderWinsOverBadQuery 坏 query + 好 header ⇒ 用 header 成功（修复优先级 bug）
func TestJWTValidatorHeaderWinsOverBadQuery(t *testing.T) {
	engine := newJWTTestEngine("header-token")
	cases := map[string]string{
		"普通 API": "/api/tasks?token=bad-token",
		"头像":     "/static/uploads/avatars/1.png?token=bad-token",
		"SSE":    "/api/sse/reminders?token=bad-token",
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			status, code := doJWTRequest(t, engine, target, "Bearer header-token")
			if status != http.StatusOK || code != 0 {
				t.Fatalf("%s 应以 header 通过，实际 status=%d code=%d", name, status, code)
			}
		})
	}
}

// TestJWTValidatorNoCredentials 无凭证 ⇒ 10041
func TestJWTValidatorNoCredentials(t *testing.T) {
	engine := newJWTTestEngine("query-token")
	status, code := doJWTRequest(t, engine, "/api/tasks", "")
	if status != http.StatusOK || code != 10041 {
		t.Fatalf("无凭证应 10041，实际 status=%d code=%d", status, code)
	}
}

// TestJWTValidatorPrefixBoundary 前缀边界：/api/ssevil 与 /static/uploads/avatars-export
// 不得被允许清单误放行
func TestJWTValidatorPrefixBoundary(t *testing.T) {
	engine := newJWTTestEngine("query-token")
	for _, target := range []string{
		"/api/ssevil?token=query-token",
		"/static/uploads/avatars-export?token=query-token",
	} {
		status, code := doJWTRequest(t, engine, target, "")
		if status == http.StatusOK && code == 0 {
			t.Fatalf("%s 不得因前缀被放行", target)
		}
		if code != 10041 {
			t.Fatalf("%s 应被拒为 10041，实际 status=%d code=%d", target, status, code)
		}
	}
}

// TestJWTValidatorTraversalNotAllowlisted 路径穿越 / 大小写 / 编码变体不得把 query token
// 带到普通 API（未匹配路由 ⇒ 404，或按普通路由判定为 10041）
func TestJWTValidatorTraversalNotAllowlisted(t *testing.T) {
	engine := newJWTTestEngine("query-token")
	for _, target := range []string{
		"/static/uploads/avatars/../tasks?token=query-token",
		"/static/uploads/avatars/%2e%2e/tasks?token=query-token",
		"/api/sse/../tasks?token=query-token",
		"/STATIC/UPLOADS/AVATARS/1.png?token=query-token",
	} {
		status, _ := doJWTRequest(t, engine, target, "")
		// 变体路径不匹配任何路由 ⇒ 404，query token 不会把请求带进任何处理器
		if status != http.StatusNotFound {
			t.Fatalf("%s 应 404（未匹配路由），实际 status=%d", target, status)
		}
	}
}

// TestJWTValidatorHeaderOnAvatar 头像同样支持 Authorization（无 query 亦可）
func TestJWTValidatorHeaderOnAvatar(t *testing.T) {
	engine := newJWTTestEngine("header-token")
	status, code := doJWTRequest(
		t, engine, "/static/uploads/avatars/1.png", "Bearer header-token",
	)
	if status != http.StatusOK || code != 0 {
		t.Fatalf("头像应支持 header，实际 status=%d code=%d", status, code)
	}
}
