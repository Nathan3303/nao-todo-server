package valueobjects

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domerr "naotodoserver/domain/errors"
)

// readAvatarFixture 读取 testdata 下的真实图片样本
// @param name 文件名
// @return 文件字节
func readAvatarFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "avatar", name))
	if err != nil {
		t.Fatalf("读取样本 %s 失败: %v", name, err)
	}
	return data
}

// TestValidateAvatarImageAllowsRealImages 合法真实样本（含三种 WebP 变体）⇒ 通过且宽高识别正确
func TestValidateAvatarImageAllowsRealImages(t *testing.T) {
	cases := []struct {
		file   string
		ext    string
		format AvatarFormat
		width  int
		height int
	}{
		{"sample.png", ".png", AvatarFormatPNG, 4, 3},
		{"sample_max.png", ".png", AvatarFormatPNG, 4096, 1},
		{"sample.jpg", ".jpg", AvatarFormatJPEG, 4, 3},
		{"sample.jpg", ".jpeg", AvatarFormatJPEG, 4, 3},
		{"sample_vp8.webp", ".webp", AvatarFormatWebP, 4, 3},  // 有损 VP8
		{"sample_vp8l.webp", ".webp", AvatarFormatWebP, 4, 3}, // 无损 VP8L
		{"sample_vp8x.webp", ".webp", AvatarFormatWebP, 4, 3}, // 扩展 VP8X（含 alpha）
	}
	for _, tc := range cases {
		t.Run(tc.file+" "+tc.ext, func(t *testing.T) {
			data := readAvatarFixture(t, tc.file)
			format, err := ValidateAvatarImage(tc.ext, data)
			if err != nil {
				t.Fatalf("合法样本被拒: %v", err)
			}
			if format != tc.format {
				t.Fatalf("格式 = %q, want %q", format, tc.format)
			}
			if _, width, height, err := sniffAvatarImage(data); err != nil ||
				width != tc.width || height != tc.height {
				t.Fatalf("宽高 = %dx%d (err=%v), want %dx%d",
					width, height, err, tc.width, tc.height)
			}
		})
	}
}

// TestValidateAvatarImageRejects 拒绝矩阵（SEC-08）：
// 伪 PNG（文本/zip 改名）、SVG/GIF（含改名伪装）、扩展名与内容不一致、超尺寸
func TestValidateAvatarImageRejects(t *testing.T) {
	cases := []struct {
		name string
		ext  string
		data []byte
	}{
		{"文本改名 png", ".png", readAvatarFixture(t, "fake.txt")},
		{"zip 改名 png", ".png", readAvatarFixture(t, "fake.zip")},
		{"SVG 用 svg 扩展", ".svg", readAvatarFixture(t, "sample.svg")},
		{"SVG 改名 png", ".png", readAvatarFixture(t, "sample.svg")},
		{"GIF 用 gif 扩展", ".gif", readAvatarFixture(t, "sample.gif")},
		{"GIF 改名 png", ".png", readAvatarFixture(t, "sample.gif")},
		{"真 PNG 用 jpg 扩展", ".jpg", readAvatarFixture(t, "sample.png")},
		{"真 JPEG 用 webp 扩展", ".webp", readAvatarFixture(t, "sample.jpg")},
		{"真 WebP 用 png 扩展", ".png", readAvatarFixture(t, "sample_vp8.webp")},
		{"超出最大边长", ".png", readAvatarFixture(t, "sample_wide.png")},
		{"空文件", ".png", nil},
		{"截断的 PNG 头", ".png", readAvatarFixture(t, "sample.png")[:10]},
		{"截断的 JPEG", ".jpg", readAvatarFixture(t, "sample.jpg")[:12]},
		{"截断的 WebP", ".webp", readAvatarFixture(t, "sample_vp8.webp")[:18]},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateAvatarImage(tc.ext, tc.data)
			if err == nil {
				t.Fatalf("应被拒绝（ext=%s, %d 字节）", tc.ext, len(tc.data))
			}
			if !domerr.IsPublicError(err) {
				t.Fatalf("拒绝原因必须可回显（PublicError），实际: %T %v", err, err)
			}
			for _, leak := range []string{"/", "testdata", "png:", "RIFF", "IHDR"} {
				if strings.Contains(err.Error(), leak) && leak != "/" {
					t.Fatalf("文案泄漏内部细节 %q: %s", leak, err)
				}
			}
		})
	}
}

// TestValidateAvatarImageWebPFailsClosedOnUnknownVariant 无法确定宽高的 WebP 变体 ⇒ 拒绝
func TestValidateAvatarImageWebPFailsClosedOnUnknownVariant(t *testing.T) {
	data := readAvatarFixture(t, "sample_vp8.webp")
	// 把 chunk 四字符改成未知变体（ALPH）：仍满足 RIFF/WEBP 但无法取画布尺寸
	mutated := append([]byte(nil), data...)
	copy(mutated[12:16], "ALPH")
	_, err := ValidateAvatarImage(".webp", mutated)
	if err == nil {
		t.Fatal("未知 WebP 变体必须拒绝（fail-closed）")
	}
	if !errors.Is(err, errAvatarFormat) {
		t.Fatalf("未知变体应返回格式错误，实际: %v", err)
	}
}

// TestValidateAvatarURL 允许/拒绝矩阵：scheme、长度、控制字符、userinfo、私网 IP 字面量
func TestValidateAvatarURL(t *testing.T) {
	allowed := []string{
		"https://cdn.example.com/a.png",
		"http://example.com/a.jpg",
		"https://example.com:8443/path/to/a.webp?x=1",
		"https://8.8.8.8/a.png",
	}
	for _, raw := range allowed {
		if err := ValidateAvatarURL(raw); err != nil {
			t.Fatalf("合法 URL 被拒 %q: %v", raw, err)
		}
	}

	//nolint:gosec // G101: 以下均为"必须被拒绝"的测试样本，非真实凭据
	rejected := map[string]string{
		"空":          "",
		"javascript": "javascript:alert(1)",
		"data":       "data:image/png;base64,AAAA",
		"file":       "file:///etc/passwd",
		"协议相对":       "//example.com/a.png",
		"无 scheme":   "example.com/a.png",
		"ftp":        "ftp://example.com/a.png",
		"带回环 IPv4":   "http://127.0.0.1/a.png",
		"回环 127.1":   "http://127.0.0.2/a.png",
		"私网 10":      "http://10.1.2.3/a.png",
		"私网 172.16":  "http://172.16.0.9/a.png",
		"私网 192.168": "http://192.168.1.10/a.png",
		"链路本地":       "http://169.254.169.254/latest/meta-data",
		"CGNAT":      "http://100.64.0.1/a.png",
		"未指定":        "http://0.0.0.0/a.png",
		"IPv6 回环":    "http://[::1]/a.png",
		"IPv6 ULA":   "http://[fd00::1]/a.png",
		"IPv6 链路本地":  "http://[fe80::1]/a.png",
		"localhost":  "http://localhost/a.png",
		"带 userinfo": "https://user:pass@example.com/a.png",
		"含空格":        "https://example.com/a b.png",
		"含换行":        "https://example.com/a.png\nHost: evil",
		"超长":         "https://example.com/" + strings.Repeat("a", AvatarURLMaxLen),
	}
	for name, raw := range rejected {
		t.Run(name, func(t *testing.T) {
			err := ValidateAvatarURL(raw)
			if err == nil {
				t.Fatalf("应被拒绝: %q", raw)
			}
			if !domerr.IsPublicError(err) {
				t.Fatalf("拒绝原因必须可回显，实际 %T", err)
			}
		})
	}
}
