package user_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	userDto "naotodoserver/application/user/dto"
	domerr "naotodoserver/domain/errors"
)

// TestUpdateAvatarByFileContentMatrix 内容校验矩阵（SEC-08）：
// 伪 PNG（文本/zip 改名）、SVG/GIF（含改名伪装）、扩展名与内容不一致 ⇒ 拒绝且不落盘
func TestUpdateAvatarByFileContentMatrix(t *testing.T) {
	cases := []struct {
		name     string
		file     string // testdata 样本
		filename string // 上传文件名（决定扩展名）
	}{
		{"文本改名 png", "fake.txt", "me.png"},
		{"zip 改名 png", "fake.zip", "me.png"},
		{"SVG 用 svg 扩展", "sample.svg", "me.svg"},
		{"SVG 改名 png", "sample.svg", "me.png"},
		{"GIF 用 gif 扩展", "sample.gif", "me.gif"},
		{"GIF 改名 png", "sample.gif", "me.png"},
		{"真 PNG 用 jpg 扩展", "sample.png", "me.jpg"},
		{"真 JPEG 用 png 扩展", "sample.jpg", "me.png"},
		{"超最大边长", "sample_wide.png", "me.png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &fakeAvatarUserRepo{}
			app, avatarDir := newAvatarTestApp(t, userRepo, &fakeCommentApp{})
			data := readAvatarFixture(t, tc.file)

			_, err := app.UpdateAvatarByFile(
				context.Background(), 42, bytes.NewReader(data), tc.filename, int64(len(data)),
			)
			if err == nil {
				t.Fatalf("必须拒绝（%s 作为 %s）", tc.file, tc.filename)
			}
			if !domerr.IsPublicError(err) {
				t.Fatalf("拒绝原因必须可回显，实际: %T %v", err, err)
			}
			assertNoAvatarFileLeak(t, err, avatarDir, userRepo)
		})
	}
}

// TestUpdateAvatarByFileWebPVariants 三种真实 WebP 变体均通过
func TestUpdateAvatarByFileWebPVariants(t *testing.T) {
	for _, sample := range []string{
		"sample_vp8.webp",  // VP8 有损
		"sample_vp8l.webp", // VP8L 无损
		"sample_vp8x.webp", // VP8X 扩展（含 alpha）
	} {
		t.Run(sample, func(t *testing.T) {
			userRepo := &fakeAvatarUserRepo{}
			app, avatarDir := newAvatarTestApp(t, userRepo, &fakeCommentApp{})
			data := readAvatarFixture(t, sample)

			out, err := app.UpdateAvatarByFile(
				context.Background(), 42, bytes.NewReader(data), "me.webp", int64(len(data)),
			)
			if err != nil {
				t.Fatalf("合法 WebP 被拒: %v", err)
			}
			if !strings.HasSuffix(out.AvatarURL, ".webp") {
				t.Fatalf("返回路径 %q 应以 .webp 结尾", out.AvatarURL)
			}
			if _, err := os.Stat(filepath.Join(avatarDir, filepath.Base(out.AvatarURL))); err != nil {
				t.Fatalf("WebP 文件应落盘: %v", err)
			}
		})
	}
}

// TestUpdateAvatarByFileUsesActualSize 不信声明 size：声明很小但实际超限 ⇒ 拒绝且清理
func TestUpdateAvatarByFileUsesActualSize(t *testing.T) {
	userRepo := &fakeAvatarUserRepo{}
	app, avatarDir := newAvatarTestApp(t, userRepo, &fakeCommentApp{})
	// 构造超过 2MB 的真实 PNG（尾部补零不影响魔数/尺寸解析，但实际字节数超限）
	base := readAvatarFixture(t, "sample.png")
	oversize := make([]byte, 2<<20+1)
	copy(oversize, base)

	_, err := app.UpdateAvatarByFile(
		context.Background(), 42, bytes.NewReader(oversize), "me.png", 10, // 声明 10 字节
	)
	if err == nil || !domerr.IsPublicError(err) {
		t.Fatalf("实际超限必须拒绝，实际: %v", err)
	}
	if !strings.Contains(err.Error(), "文件大小超出限制") {
		t.Fatalf("文案不符: %v", err)
	}
	assertNoAvatarFileLeak(t, err, avatarDir, userRepo)
}

// TestUpdateAvatarURLValidation URL 方式矩阵（SEC-08）：scheme/私网/长度 ⇒ 拒绝且不写库
func TestUpdateAvatarURLValidation(t *testing.T) {
	rejected := map[string]string{
		"javascript": "javascript:alert(1)",
		"file":       "file:///etc/passwd",
		"data":       "data:image/png;base64,AAAA",
		"私网":         "http://192.168.1.10/a.png",
		"回环":         "http://127.0.0.1/a.png",
		"元数据":        "http://169.254.169.254/latest/meta-data",
		"超长":         "https://example.com/" + strings.Repeat("a", 300),
	}
	for name, raw := range rejected {
		t.Run(name, func(t *testing.T) {
			userRepo := &fakeAvatarUserRepo{}
			app, _ := newAvatarTestApp(t, userRepo, &fakeCommentApp{})

			_, err := app.UpdateAvatar(context.Background(), 42, userDto.UpdateAvatarInput{AvatarURL: raw})
			if err == nil || !domerr.IsPublicError(err) {
				t.Fatalf("应拒绝且可回显，实际: %v", err)
			}
			if userRepo.avatar != "" {
				t.Fatalf("校验失败不得写库，实际 %q", userRepo.avatar)
			}
		})
	}

	// 合法 https URL：原样（trim 后）写库
	userRepo := &fakeAvatarUserRepo{}
	app, _ := newAvatarTestApp(t, userRepo, &fakeCommentApp{})
	valid := "  https://cdn.example.com/a.png  "
	if _, err := app.UpdateAvatar(context.Background(), 42, userDto.UpdateAvatarInput{AvatarURL: valid}); err != nil {
		t.Fatalf("合法 URL 应通过: %v", err)
	}
	if userRepo.avatar != "https://cdn.example.com/a.png" {
		t.Fatalf("写库值 = %q, want trim 后的 URL", userRepo.avatar)
	}
}

// assertNoAvatarFileLeak 断言拒绝时：无文件落盘、未改用户表、文案不含路径
func assertNoAvatarFileLeak(
	t *testing.T, err error, avatarDir string, userRepo *fakeAvatarUserRepo,
) {
	t.Helper()
	if entries, readErr := os.ReadDir(avatarDir); readErr == nil && len(entries) != 0 {
		t.Fatalf("校验失败不应产生文件，实际 %d 个", len(entries))
	}
	if userRepo.avatar != "" {
		t.Fatalf("校验失败不应更新头像，实际 %q", userRepo.avatar)
	}
	for _, leak := range []string{avatarDir, "testdata", "avatars/", "../"} {
		if strings.Contains(err.Error(), leak) {
			t.Fatalf("文案不得包含路径 %q: %v", leak, err)
		}
	}
}
