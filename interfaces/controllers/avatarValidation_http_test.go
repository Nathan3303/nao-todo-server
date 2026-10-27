package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/storage"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// avatarRepoStub 头像路径所需的最小用户仓储
type avatarRepoStub struct {
	repositories.User
	avatar string
}

func (s *avatarRepoStub) FindById(
	_ context.Context, id domaintypes.UserID,
) (*entities.User, error) {
	return &entities.User{
		EntityBase: domaintypes.EntityBase{Id: int64(id)},
		Avatar:     s.avatar,
	}, nil
}

func (s *avatarRepoStub) UpdateAvatar(
	_ context.Context, _ domaintypes.UserID, avatarURL string,
) error {
	s.avatar = avatarURL
	return nil
}

// avatarCommentStub 评论资料同步（上传成功后会调用）
type avatarCommentStub struct {
	taskApp.TaskCommentApp
}

func (s *avatarCommentStub) SyncTaskCommentUserProfile(
	_ context.Context, _ int64, _, _ string,
) error {
	return nil
}

// newAvatarUploadEngine 真实控制器 + 真实本地存储 + 假仓储；返回引擎与头像目录
func newAvatarUploadEngine(t *testing.T) (*gin.Engine, *avatarRepoStub, string) {
	t.Helper()
	uploadDir := t.TempDir()
	old := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		UploadDir:   uploadDir,
		AvatarDir:   "avatars",
		StaticPath:  "/static/uploads",
		MaxFileSize: 2 << 20,
	}}
	t.Cleanup(func() { conf.Conf = old })

	repo := &avatarRepoStub{}
	app := userApp.NewUserApp(repo, nil, &avatarCommentStub{}, storage.NewAvatarStorage())
	ctrl := controllers.NewUserController(app)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.PUT("/api/user/avatar", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.UpdateUserAvatar(ctx)
	})
	return engine, repo, filepath.Join(uploadDir, "avatars")
}

// avatarFixture 读取真实图片样本（与 domain 层共用 testdata）
func avatarFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(
		"..", "..", "domain", "identity", "valueobjects", "testdata", "avatar", name,
	))
	if err != nil {
		t.Fatalf("读取样本 %s 失败: %v", name, err)
	}
	return data
}

// avatarRes 头像接口响应
type avatarRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// postAvatarMultipart 以 multipart 上传
func postAvatarMultipart(t *testing.T, engine *gin.Engine, filename string, content []byte, withField bool) avatarRes {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if withField {
		part, err := writer.CreateFormFile("avatar", filename)
		if err != nil {
			t.Fatalf("构造 multipart 失败: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("写入 multipart 失败: %v", err)
		}
	} else {
		if err := writer.WriteField("note", "no-file"); err != nil {
			t.Fatalf("写入字段失败: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("关闭 multipart 失败: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/user/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	res := avatarRes{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP 状态 = %d, want 200", rec.Code)
	}
	return res
}

// postAvatarJSON 以 URL 方式上传
func postAvatarJSON(t *testing.T, engine *gin.Engine, payload string) avatarRes {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/user/avatar", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	res := avatarRes{}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	return res
}

// TestAvatarUploadContentMatrixHTTP 上传矩阵（SEC-08）：伪 PNG/SVG/GIF ⇒ 10082 + 统一文案、无路径泄漏
func TestAvatarUploadContentMatrixHTTP(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		filename string
	}{
		{"文本改名 png", "fake.txt", "me.png"},
		{"zip 改名 png", "fake.zip", "me.png"},
		{"SVG 改名 png", "sample.svg", "me.png"},
		{"GIF 改名 png", "sample.gif", "me.png"},
		{"PNG 用 jpg 扩展", "sample.png", "me.jpg"},
		{"超最大边长", "sample_wide.png", "me.png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, repo, avatarDir := newAvatarUploadEngine(t)
			res := postAvatarMultipart(t, engine, tc.filename, avatarFixture(t, tc.file), true)

			if res.Code != 10082 {
				t.Fatalf("code = %d, want 10082", res.Code)
			}
			if res.Message != "更新用户头像失败" {
				t.Fatalf("message = %q, want 固定文案", res.Message)
			}
			for _, leak := range []string{"testdata", "avatars/", "/etc/", "\\\\", "multipart"} {
				if strings.Contains(res.Error, leak) {
					t.Fatalf("error 泄漏内部细节 %q: %q", leak, res.Error)
				}
			}
			if res.Error == "" {
				t.Fatal("error 必须给出可读文案")
			}
			if repo.avatar != "" {
				t.Fatalf("拒绝时不得写库，实际 %q", repo.avatar)
			}
			if entries, err := os.ReadDir(avatarDir); err == nil && len(entries) != 0 {
				t.Fatalf("拒绝时不得落盘，实际 %d 个文件", len(entries))
			}
		})
	}
}

// TestAvatarUploadAcceptsRealFormatsHTTP 合法 PNG / JPEG / WebP(VP8/VP8L/VP8X) ⇒ 10080
func TestAvatarUploadAcceptsRealFormatsHTTP(t *testing.T) {
	cases := []struct {
		file     string
		filename string
	}{
		{"sample.png", "me.png"},
		{"sample.jpg", "me.jpg"},
		{"sample_vp8.webp", "me.webp"},
		{"sample_vp8l.webp", "me.webp"},
		{"sample_vp8x.webp", "me.webp"},
	}
	for _, tc := range cases {
		t.Run(tc.filename+" "+tc.file, func(t *testing.T) {
			engine, repo, _ := newAvatarUploadEngine(t)
			res := postAvatarMultipart(t, engine, tc.filename, avatarFixture(t, tc.file), true)
			if res.Code != 10080 {
				t.Fatalf("code = %d, want 10080（error=%q）", res.Code, res.Error)
			}
			if !strings.HasPrefix(repo.avatar, "/static/uploads/avatars/42-") {
				t.Fatalf("用户表头像 = %q, want 随机名相对路径", repo.avatar)
			}
		})
	}
}

// TestAvatarUploadMissingFieldNoLeak 缺少文件字段 ⇒ 10082 + 固定文案（不回显 multipart 内部错误）
func TestAvatarUploadMissingFieldNoLeak(t *testing.T) {
	engine, _, _ := newAvatarUploadEngine(t)
	res := postAvatarMultipart(t, engine, "", nil, false)
	if res.Code != 10082 {
		t.Fatalf("code = %d, want 10082", res.Code)
	}
	if res.Message != "更新用户头像失败" {
		t.Fatalf("message = %q", res.Message)
	}
	for _, leak := range []string{"multipart", "http:", "no such file", "boundary"} {
		if strings.Contains(res.Error, leak) {
			t.Fatalf("error 泄漏内部细节 %q: %q", leak, res.Error)
		}
	}
}

// TestAvatarURLValidationHTTP URL 方式：javascript:/file:/私网 ⇒ 10082；合法 https ⇒ 10080
func TestAvatarURLValidationHTTP(t *testing.T) {
	rejected := map[string]string{
		"javascript": `{"avatarURL":"javascript:alert(1)"}`,
		"file":       `{"avatarURL":"file:///etc/passwd"}`,
		"private":    `{"avatarURL":"http://10.0.0.5/a.png"}`,
		"loopback":   `{"avatarURL":"http://127.0.0.1/a.png"}`,
		"metadata":   `{"avatarURL":"http://169.254.169.254/latest/meta-data"}`,
	}
	for name, payload := range rejected {
		t.Run(name, func(t *testing.T) {
			engine, repo, _ := newAvatarUploadEngine(t)
			res := postAvatarJSON(t, engine, payload)
			if res.Code != 10082 {
				t.Fatalf("code = %d, want 10082", res.Code)
			}
			if res.Message != "更新用户头像失败" || !strings.Contains(res.Error, "头像链接不合法") {
				t.Fatalf("文案不符: message=%q error=%q", res.Message, res.Error)
			}
			if repo.avatar != "" {
				t.Fatalf("拒绝时不得写库，实际 %q", repo.avatar)
			}
		})
	}

	engine, repo, _ := newAvatarUploadEngine(t)
	res := postAvatarJSON(t, engine, `{"avatarURL":"https://cdn.example.com/a.png"}`)
	if res.Code != 10080 {
		t.Fatalf("合法 URL code = %d, want 10080（error=%q）", res.Code, res.Error)
	}
	if repo.avatar != "https://cdn.example.com/a.png" {
		t.Fatalf("写库值 = %q", repo.avatar)
	}
}
