package controllers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authApp "naotodoserver/application/auth"
	userApp "naotodoserver/application/user"
	"naotodoserver/conf"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/storage"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// fakeAvatarAuth 只实现头像路由用到的 Validate / RateLimit
type fakeAvatarAuth struct {
	authApp.AuthApp
	validToken   string
	rateLimitErr error
}

func (f *fakeAvatarAuth) Validate(
	_ context.Context, token string,
) (domaintypes.UserID, error) {
	if token == f.validToken {
		return domaintypes.UserID(42), nil
	}
	return 0, errors.New("invalid token")
}

func (f *fakeAvatarAuth) RateLimit(_ context.Context, _ string, _ int64) error {
	return f.rateLimitErr
}

// newAvatarHTTPEngine 按线上挂载顺序构建头像路由（JWT → 用户限流 → 控制器），
// 头像存储用真实本地实现（临时目录），其余依赖不参与。
func newAvatarHTTPEngine(t *testing.T, auth *fakeAvatarAuth) (*gin.Engine, string) {
	t.Helper()
	uploadDir := t.TempDir()
	old := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		UploadDir:  uploadDir,
		AvatarDir:  "avatars",
		StaticPath: "/static/uploads",
	}}
	t.Cleanup(func() { conf.Conf = old })

	avatarDir := filepath.Join(uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		t.Fatalf("创建头像目录失败: %v", err)
	}

	app := userApp.NewUserApp(nil, nil, nil, storage.NewAvatarStorage())
	ctrl := controllers.NewUserController(app)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET(
		"/static/uploads/avatars/:filename",
		middlewares.JWTValidator(auth),
		middlewares.UserRateLimiter(auth, 120, "avatar"),
		ctrl.GetAvatar,
	)
	return engine, avatarDir
}

// getAvatar 请求头像，返回 HTTP 状态 / 业务 code / 响应体
func getAvatar(
	t *testing.T, engine *gin.Engine, target, authHeader string,
) (int, int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var body struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code, rec.Body.String()
}

// TestGetAvatarLegacyFilenameReadable 旧 {userId}.ext 仍可读（兼容性自证）
func TestGetAvatarLegacyFilenameReadable(t *testing.T) {
	auth := &fakeAvatarAuth{validToken: "good-token"}
	engine, avatarDir := newAvatarHTTPEngine(t, auth)
	if err := os.WriteFile(filepath.Join(avatarDir, "42.png"), []byte("legacy"), 0o600); err != nil {
		t.Fatalf("写入旧头像失败: %v", err)
	}

	status, _, body := getAvatar(
		t, engine, "/static/uploads/avatars/42.png", "Bearer good-token",
	)
	if status != http.StatusOK || body != "legacy" {
		t.Fatalf("旧头像应可读，实际 status=%d body=%q", status, body)
	}
}

// TestGetAvatarQueryTokenStillAllowed 头像仍允许 ?token=（SEC-06 允许清单端到端）
func TestGetAvatarQueryTokenStillAllowed(t *testing.T) {
	auth := &fakeAvatarAuth{validToken: "good-token"}
	engine, avatarDir := newAvatarHTTPEngine(t, auth)
	if err := os.WriteFile(filepath.Join(avatarDir, "42.png"), []byte("legacy"), 0o600); err != nil {
		t.Fatalf("写入旧头像失败: %v", err)
	}

	status, _, body := getAvatar(
		t, engine, "/static/uploads/avatars/42.png?token=good-token", "",
	)
	if status != http.StatusOK || body != "legacy" {
		t.Fatalf("头像 query token 应可用，实际 status=%d body=%q", status, body)
	}
}

// TestGetAvatarNonexistentRandomName 不存在的随机名 ⇒ 404（枚举探测拿不到 200/404 差异以外信息）
func TestGetAvatarNonexistentRandomName(t *testing.T) {
	auth := &fakeAvatarAuth{validToken: "good-token"}
	engine, _ := newAvatarHTTPEngine(t, auth)

	target := "/static/uploads/avatars/42-0123456789abcdef0123456789abcdef.png"
	status, code, _ := getAvatar(t, engine, target, "Bearer good-token")
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
	if code != 10084 {
		t.Fatalf("code = %d, want 10084", code)
	}
}

// TestGetAvatarRateLimited 限流触发 ⇒ 10051（阈值 By user 由 UserRateLimiter 保证）
func TestGetAvatarRateLimited(t *testing.T) {
	auth := &fakeAvatarAuth{
		validToken:   "good-token",
		rateLimitErr: errors.New("超过配额"),
	}
	engine, avatarDir := newAvatarHTTPEngine(t, auth)
	if err := os.WriteFile(filepath.Join(avatarDir, "42.png"), []byte("legacy"), 0o600); err != nil {
		t.Fatalf("写入旧头像失败: %v", err)
	}

	status, code, _ := getAvatar(
		t, engine, "/static/uploads/avatars/42.png", "Bearer good-token",
	)
	if status != http.StatusOK || code != 10051 {
		t.Fatalf("超限应 10051，实际 status=%d code=%d", status, code)
	}
}

// TestGetAvatarInvalidToken 凭证无效 ⇒ 10041（先鉴权后限流）
func TestGetAvatarInvalidToken(t *testing.T) {
	auth := &fakeAvatarAuth{validToken: "good-token"}
	engine, _ := newAvatarHTTPEngine(t, auth)

	status, code, _ := getAvatar(
		t, engine, "/static/uploads/avatars/42.png", "Bearer bad-token",
	)
	if status != http.StatusOK || code != 10041 {
		t.Fatalf("无效凭证应 10041，实际 status=%d code=%d", status, code)
	}
}

// TestGetAvatarContentType 成功响应仍带 image/png 与 immutable 缓存头（客户端零改动）
func TestGetAvatarContentType(t *testing.T) {
	auth := &fakeAvatarAuth{validToken: "good-token"}
	engine, avatarDir := newAvatarHTTPEngine(t, auth)
	if err := os.WriteFile(filepath.Join(avatarDir, "42.png"), []byte("png"), 0o600); err != nil {
		t.Fatalf("写入头像失败: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet, "/static/uploads/avatars/42.png", nil,
	)
	req.Header.Set("Authorization", "Bearer good-token")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "image/png") {
		t.Fatalf("Content-Type = %q, want image/png", got)
	}
	if got := rec.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("Cache-Control = %q, 应保留 immutable", got)
	}
}
