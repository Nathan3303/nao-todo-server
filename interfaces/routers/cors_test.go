// T520 / DEF-80 服务端 CORS 放行规则单测。
//
// 覆盖 PRD 2026-10-22-server-cors-origins §七：
// AC1 既有正式域零回归 · AC2 伪装来源全拒 · AC3 局域网仅 debug 生效、预览域与 debug 无关
// · AC4 预检语义（精确回显 / allow-credentials / 被拒来源无 ACAO）· AC6 矩阵全绿。
// 仅用 httptest + gin，不依赖 MySQL / Redis。
package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsAllowedOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		debug  bool
		want   bool
	}{
		// AC1：既有 4 条精确来源，debug 两态都必须放行（不变量 4）
		{"精确-本地5173", "http://localhost:5173", false, true},
		{"精确-本地5174", "http://localhost:5174", false, true},
		{"精确-正式域", "https://todo.nathanao.space", false, true},
		{"精确-正式域be", "https://todobe.nathanao.space", false, true},
		{"精确-正式域-debug开", "https://todobe.nathanao.space", true, true},

		// 预览域：与 debug 无关（AC3）
		{"预览-分支域名-debug关", "https://naotodo-git-feat-217-abc-nathan3303s-projects.vercel.app", false, true},
		{"预览-分支域名-debug开", "https://naotodo-git-feat-217-abc-nathan3303s-projects.vercel.app", true, true},

		// AC2：伪装来源一律拒绝
		{"伪装-后缀拼接", "https://naotodo-x-nathan3303s-projects.vercel.app.evil.com", true, false},
		{"伪装-前缀伪造", "https://evil-vercel.app", true, false},
		{"伪装-其他team", "https://naotodo-x-someoneelse-projects.vercel.app", true, false},
		{"伪装-本项目前缀但无team后缀", "https://naotodo-x-nathan3303s-projects.vercel.app.evil", true, false},
		{"伪装-http协议", "http://naotodo-x-nathan3303s-projects.vercel.app", true, false},
		{"伪装-大写域名", "https://NAOTODO-X-nathan3303s-projects.vercel.app", true, false},
		{"伪装-末尾空格", "https://naotodo-x-nathan3303s-projects.vercel.app ", true, false},

		// AC3：局域网网段仅 debug 生效
		{"局域网-192.168-debug开", "http://192.168.2.211:5173", true, true},
		{"局域网-192.168-debug关", "http://192.168.2.211:5173", false, false},
		{"局域网-10-debug开", "http://10.0.0.9:5174", true, true},
		{"局域网-10-debug关", "http://10.0.0.9:5174", false, false},
		{"局域网-172.16-debug开", "http://172.16.0.1:5173", true, true},
		{"局域网-172.31-debug开", "http://172.31.255.254:5174", true, true},
		{"局域网-172.15-越界", "http://172.15.0.1:5173", true, false},
		{"局域网-172.32-越界", "http://172.32.0.1:5173", true, false},

		// 端口白名单
		{"端口-9999拒绝", "http://192.168.2.211:9999", true, false},
		{"端口-5180拒绝", "http://192.168.2.211:5180", true, false},
		{"端口-缺省80拒绝", "http://192.168.2.211", true, false},

		// 其他负例
		{"负例-公网IP", "http://8.8.8.8:5173", true, false},
		{"负例-https局域网", "https://192.168.2.211:5173", true, false},
		{"负例-空origin", "", true, false},
		{"负例-未知域名", "https://foo.example.com", true, false},
		{"负例-localhost端口越界", "http://localhost:3000", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedOrigin(tt.origin, tt.debug); got != tt.want {
				t.Fatalf("isAllowedOrigin(%q, debug=%v) = %v, want %v", tt.origin, tt.debug, got, tt.want)
			}
		})
	}
}

// corsEngine 构造只挂 CORS 中间件的最小引擎（不触达控制器 / 仓储）。
func corsEngine(debug bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(newCORSMiddleware(debug))
	engine.GET("/api/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})
	engine.POST("/api/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})
	return engine
}

// preflight 发起一次 OPTIONS 预检。
func preflight(engine *gin.Engine, origin string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodOptions, "/api/ping", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	return recorder
}

// TestCORSPreflightMatrix AC1/AC2/AC3/AC4：预检响应头的精确语义。
func TestCORSPreflightMatrix(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		debug  bool
		allow  bool
	}{
		{"正式域放行", "https://todobe.nathanao.space", false, true},
		{"本地端口放行", "http://localhost:5173", false, true},
		{"预览域放行-debug关", "https://naotodo-git-feat-52-cors-nathan3303s-projects.vercel.app", false, true},
		{"局域网放行-debug开", "http://192.168.2.211:5173", true, true},
		{"局域网拒绝-debug关", "http://192.168.2.211:5173", false, false},
		{"伪装后缀拒绝", "https://naotodo-x-nathan3303s-projects.vercel.app.evil.com", true, false},
		{"伪装裸域拒绝", "https://evil-vercel.app", true, false},
		{"越权端口拒绝", "http://192.168.2.211:9999", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := preflight(corsEngine(tt.debug), tt.origin)
			acao := recorder.Header().Get("Access-Control-Allow-Origin")

			if tt.allow {
				if recorder.Code != http.StatusNoContent {
					t.Fatalf("预检状态码 = %d, want %d", recorder.Code, http.StatusNoContent)
				}
				if acao != tt.origin {
					t.Fatalf("ACAO = %q, want 精确回显 %q", acao, tt.origin)
				}
				if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
					t.Fatalf("allow-credentials = %q, want true", got)
				}
				if got := recorder.Header().Get("Access-Control-Allow-Methods"); got == "" {
					t.Fatal("放行来源的预检缺少 access-control-allow-methods")
				}
				if got := recorder.Header().Get("Access-Control-Allow-Headers"); got == "" {
					t.Fatal("放行来源的预检缺少 access-control-allow-headers")
				}
				return
			}

			if acao != "" {
				t.Fatalf("被拒来源不得带 ACAO，实际 = %q", acao)
			}
			if acao == "*" {
				t.Fatal("AllowCredentials=true 时不得回显 *")
			}
		})
	}
}

// TestCORSNormalRequestEcho AC4：非预检请求同样精确回显，被拒来源无 ACAO。
func TestCORSNormalRequestEcho(t *testing.T) {
	allowed := "https://todo.nathanao.space"
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("Origin", allowed)
	corsEngine(false).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("普通请求状态码 = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != allowed {
		t.Fatalf("ACAO = %q, want %q", got, allowed)
	}

	denied := "https://naotodo-x-nathan3303s-projects.vercel.app.evil.com"
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	req.Header.Set("Origin", denied)
	corsEngine(true).ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("被拒来源不得带 ACAO，实际 = %q", got)
	}
}
