package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"naotodoserver/conf"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// bodyLimitFixture 构造与生产一致的上限配置（默认 2MB / 头像 5MB）
func bodyLimitFixture(t *testing.T) {
	t.Helper()
	old := conf.Conf
	conf.Conf = &conf.Config{Server: &conf.Server{
		MaxRequestBodyBytes: 2 << 20,
		MaxUploadBodyBytes:  5 << 20,
	}}
	t.Cleanup(func() { conf.Conf = old })
}

// newBodyLimitEngine 复刻生产挂载顺序：BodyLimit(默认, bodyLimitOverrides()) → 路由
func newBodyLimitEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middlewares.BodyLimit(
		conf.Conf.Server.MaxRequestBodyBytes,
		bodyLimitOverrides(),
	))
	// 未鉴权：模拟 JWTValidator 拒绝（证明请求已越过 BodyLimit 层）
	unauthenticated := func(ctx *gin.Context) {
		controllers.Failure(ctx, types.ResponseData{Code: 10041, Message: "用户凭证验证失败"})
		ctx.Abort()
	}
	engine.PUT("/api/user/avatar", unauthenticated)
	engine.PUT("/api/projects/:projectId", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, types.ResponseData{Code: 20023, Message: "更新清单失败"})
	})
	return engine
}

// sendBody 发送指定大小的请求体，返回 HTTP 状态与业务码
func sendBody(t *testing.T, engine *gin.Engine, path string, size int) (int, int) {
	t.Helper()
	body := bytes.Repeat([]byte("a"), size)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var res struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	return rec.Code, res.Code
}

// TestBodyLimitOverridesUseRealAvatarPath F2：覆盖表必须是真实路由 /api/user/avatar
func TestBodyLimitOverridesUseRealAvatarPath(t *testing.T) {
	bodyLimitFixture(t)
	overrides := bodyLimitOverrides()

	if got, ok := overrides["/api/user/avatar"]; !ok || got != conf.Conf.Server.MaxUploadBodyBytes {
		t.Fatalf("头像路径覆盖缺失或值错误: %v（期望 %d）", overrides, conf.Conf.Server.MaxUploadBodyBytes)
	}
	if _, stale := overrides["/api/users/avatar"]; stale {
		t.Fatal("覆盖表不得再包含不存在的路径 /api/users/avatar")
	}
}

// TestBodyLimitAvatarPathAllows3MB F2 ①：3MB 发往 /api/user/avatar ⇒ 不得 413（越过上限层走到鉴权）
func TestBodyLimitAvatarPathAllows3MB(t *testing.T) {
	bodyLimitFixture(t)
	engine := newBodyLimitEngine()

	status, code := sendBody(t, engine, "/api/user/avatar", 3<<20)
	if status == http.StatusRequestEntityTooLarge {
		t.Fatalf("3MB 不应被 413 拦下（5MB 放宽未生效？），实际 status=%d", status)
	}
	if code != 10041 {
		t.Fatalf("应走到鉴权并被拒（10041），实际 status=%d code=%d", status, code)
	}
}

// TestBodyLimitOtherPathStillRejects3MB F2 ②：3MB 发往其它路径 ⇒ 仍 413
func TestBodyLimitOtherPathStillRejects3MB(t *testing.T) {
	bodyLimitFixture(t)
	engine := newBodyLimitEngine()

	status, _ := sendBody(t, engine, "/api/projects/1", 3<<20)
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("其它路径 3MB 应 413，实际 status=%d", status)
	}
}

// TestBodyLimitAvatarPathRejects6MB F2 ③：6MB 发往头像路径 ⇒ 413（放宽到 5MB，不是无限）
func TestBodyLimitAvatarPathRejects6MB(t *testing.T) {
	bodyLimitFixture(t)
	engine := newBodyLimitEngine()

	status, _ := sendBody(t, engine, "/api/user/avatar", 6<<20)
	if status != http.StatusRequestEntityTooLarge {
		t.Fatalf("6MB 应 413，实际 status=%d", status)
	}
}

// TestBodyLimitAvatarPathRejectsChunkedOversize 无 Content-Length（chunked）时仍按实际读取字节拒绝
func TestBodyLimitAvatarPathRejectsChunkedOversize(t *testing.T) {
	bodyLimitFixture(t)
	engine := newBodyLimitEngine()

	body := bytes.Repeat([]byte("a"), 6<<20)
	req := httptest.NewRequest(http.MethodPut, "/api/user/avatar", bytes.NewReader(body))
	req.ContentLength = -1 // 声明未知长度 ⇒ 走 MaxBytesReader 实际读取路径
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("chunked 6MB 应 413，实际 status=%d", rec.Code)
	}
}
