package middlewares

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	iCtx "naotodoserver/infrastructure/context"

	"github.com/gin-gonic/gin"
)

var requestIdHex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

// runRequestId 走一次中间件，返回上下文中的关联标识与响应头
func runRequestId(t *testing.T, incoming string) (string, http.Header) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestId())
	var got string
	engine.GET("/probe", func(ctx *gin.Context) {
		got = iCtx.GetRequestId(ctx.Request.Context())
		ctx.String(http.StatusOK, "ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	if incoming != "" {
		req.Header.Set("X-Request-Id", incoming)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return got, rec.Header()
}

// TestRequestIdPassesValidClientId 合法客户端标识原样透传，且不写入响应
func TestRequestIdPassesValidClientId(t *testing.T) {
	valid := "web-2026.10_27:abc" // 冒号不在白名单 ⇒ 应被丢弃
	got, _ := runRequestId(t, valid)
	if got == valid {
		t.Fatalf("非法字符的标识不应透传: %q", got)
	}
	if !requestIdHex32.MatchString(got) {
		t.Fatalf("应回退为服务端生成的 32 位 hex，实际 %q", got)
	}

	good := "web-2026.10_27-abc"
	got, header := runRequestId(t, good)
	if got != good {
		t.Fatalf("合法标识应透传: got %q want %q", got, good)
	}
	if header.Get("X-Request-Id") != "" {
		t.Fatalf("关联标识不得写入响应头，实际 %q", header.Get("X-Request-Id"))
	}
}

// TestRequestIdRejectsInjectionPayloads 日志注入载荷一律丢弃并重新生成
func TestRequestIdRejectsInjectionPayloads(t *testing.T) {
	cases := map[string]string{
		"换行注入": "abc\n{\"level\":\"error\"}",
		"空格":   "abc def",
		"超长":   repeatStr("a", 65),
		"引号":   `abc"def`,
		"路径字符": "abc/def",
		"中文":   "请求-1",
		"制表符":  "abc\tdef",
		"只有空白": "   ",
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := runRequestId(t, payload)
			if got == payload {
				t.Fatalf("注入载荷不应透传: %q", got)
			}
			if !requestIdHex32.MatchString(got) {
				t.Fatalf("应回退为服务端生成值，实际 %q", got)
			}
		})
	}
}

// TestRequestIdGeneratesWhenAbsent 缺省时服务端生成，且每次不同
func TestRequestIdGeneratesWhenAbsent(t *testing.T) {
	first, _ := runRequestId(t, "")
	second, _ := runRequestId(t, "")
	if !requestIdHex32.MatchString(first) || !requestIdHex32.MatchString(second) {
		t.Fatalf("缺省时应生成 32 位 hex: %q / %q", first, second)
	}
	if first == second {
		t.Fatal("每次请求应生成不同的关联标识")
	}
}

// repeatStr 构造重复字符串（测试用）
func repeatStr(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
