package middlewares

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// echoBodyHandler 模拟下游控制器：一次性读完整 body 并回显长度。
func echoBodyHandler(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "read error")
		return
	}
	ctx.String(http.StatusOK, "%d", len(body))
}

// testDefaultLimit 测试用全局请求体上限（2 MB，与线上默认一致）
const testDefaultLimit int64 = 2 << 20

// newBodyLimitEngine 构建带请求体上限的最小引擎（SEC-01）
func newBodyLimitEngine(overrides map[string]int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(BodyLimit(testDefaultLimit, overrides))
	engine.POST("/echo", echoBodyHandler)
	engine.PUT("/api/users/avatar", echoBodyHandler)
	engine.GET("/ping", func(ctx *gin.Context) { ctx.String(http.StatusOK, "pong") })
	return engine
}

// TestBodyLimitAllowsNormalRequest 合法请求：body 原样还原给下游（200，长度不变）
func TestBodyLimitAllowsNormalRequest(t *testing.T) {
	engine := newBodyLimitEngine(nil)
	payload := `{"tasks":[{"name":"正常请求"}]}`

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(payload))
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if want := strconv.Itoa(len(payload)); rec.Body.String() != want {
		t.Fatalf("body 长度 = %q, want %q（请求体未完整还原）", rec.Body.String(), want)
	}
}

// TestBodyLimitRejectsDeclaredOversize Content-Length 已声明 10 MB ⇒ 413（不读 body）
func TestBodyLimitRejectsDeclaredOversize(t *testing.T) {
	engine := newBodyLimitEngine(nil)
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(strings.Repeat("x", 10<<20)))
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	assertTooLarge(t, rec)
}

// TestBodyLimitRejectsChunkedOversize 未声明长度（chunked）10 MB ⇒ 413，且读取量受限于上限
func TestBodyLimitRejectsChunkedOversize(t *testing.T) {
	limit := testDefaultLimit
	engine := newBodyLimitEngine(nil)
	source := &fillerReader{remaining: 10 << 20}
	req := httptest.NewRequest(http.MethodPost, "/echo", io.NopCloser(source))
	req.ContentLength = -1 // 模拟 chunked：长度未知

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	assertTooLarge(t, rec)
	// MaxBytesReader 读到 limit+1 字节即判定超限：内存驻留量有界，不会读入完整 10 MB
	if source.read > limit+1 {
		t.Fatalf("读取量 = %d 字节, want ≤ %d（超限 body 不得整段落内存）", source.read, limit+1)
	}
}

// TestBodyLimitUploadOverride 头像上传路由单独放宽：2.5 MB 通过，其他路由 413
func TestBodyLimitUploadOverride(t *testing.T) {
	const (
		uploadLimit = 5 << 20
		bodySize    = 5 << 19 // 2.5 MB
	)
	engine := newBodyLimitEngine(map[string]int64{
		"/api/users/avatar": uploadLimit,
	})

	uploadReq := httptest.NewRequest(
		http.MethodPut, "/api/users/avatar", strings.NewReader(strings.Repeat("a", bodySize)),
	)
	uploadRec := httptest.NewRecorder()
	engine.ServeHTTP(uploadRec, uploadReq)
	if uploadRec.Code != http.StatusOK {
		t.Fatalf("头像上传 status = %d, want 200（应放宽到 %d 字节）", uploadRec.Code, uploadLimit)
	}

	otherReq := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(strings.Repeat("a", bodySize)))
	otherRec := httptest.NewRecorder()
	engine.ServeHTTP(otherRec, otherReq)
	assertTooLarge(t, otherRec)
}

// TestBodyLimitSkipsEmptyBody 无 body 的 GET ⇒ 直接放行
func TestBodyLimitSkipsEmptyBody(t *testing.T) {
	engine := newBodyLimitEngine(nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "pong" {
		t.Fatalf("GET /ping = %d %q, want 200 pong", rec.Code, rec.Body.String())
	}
}

// assertTooLarge 断言 413 + 可读错误体（SEC-01：不静默截断）
func assertTooLarge(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
	var res types.ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%q）", err, rec.Body.String())
	}
	if res.Code != http.StatusRequestEntityTooLarge || res.Message != "请求体超出上限" {
		t.Fatalf("响应 = %+v, want code 413 / message 请求体超出上限", res)
	}
}

// fillerReader 按需生成字节，避免测试自身预先分配 10 MB。
type fillerReader struct {
	remaining int64
	read      int64
}

func (r *fillerReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	n := int64(len(p))
	if n > r.remaining {
		n = r.remaining
	}
	for i := int64(0); i < n; i++ {
		p[i] = 'x'
	}
	r.remaining -= n
	r.read += n
	return int(n), nil
}
