package user

import (
	"regexp"
	"strings"
	"testing"

	"naotodoserver/conf"
)

// withAvatarUploadsConf 设置临时上传配置，返回恢复函数
func withAvatarUploadsConf(t *testing.T) {
	t.Helper()
	old := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		UploadDir:  t.TempDir(),
		AvatarDir:  "avatars",
		StaticPath: "/static/uploads",
	}}
	t.Cleanup(func() { conf.Conf = old })
}

// TestNewAvatarFilename 新文件名 = {userId}-{32 位小写 hex}.{ext}，且每次不同（不可枚举）
func TestNewAvatarFilename(t *testing.T) {
	pattern := regexp.MustCompile(`^42-[0-9a-f]{32}\.png$`)
	seen := make(map[string]bool, 20)
	for i := 0; i < 20; i++ {
		name, err := newAvatarFilename(42, ".png")
		if err != nil {
			t.Fatalf("生成文件名失败: %v", err)
		}
		if !pattern.MatchString(name) {
			t.Fatalf("文件名 %q 不符合 {userId}-{32hex}.png", name)
		}
		if seen[name] {
			t.Fatalf("随机名重复: %q", name)
		}
		seen[name] = true
	}
}

// TestLocalAvatarFilename 仅识别本服务头像目录内的相对路径（旧名/新名均可）
func TestLocalAvatarFilename(t *testing.T) {
	withAvatarUploadsConf(t)

	cases := []struct {
		name   string
		avatar string
		want   string
	}{
		{"旧文件名", "/static/uploads/avatars/42.png", "42.png"},
		{"新文件名", "/static/uploads/avatars/42-" + strings.Repeat("a", 32) + ".jpg",
			"42-" + strings.Repeat("a", 32) + ".jpg"},
		{"外部 URL", "https://cdn.example.com/avatars/42.png", ""},
		{"其它相对路径", "/uploads/42.png", ""},
		{"路径分隔符", "/static/uploads/avatars/../../42.png", ""},
		{"空值", "", ""},
		{"目录本身", "/static/uploads/avatars/", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := localAvatarFilename(c.avatar); got != c.want {
				t.Fatalf("localAvatarFilename(%q) = %q, want %q", c.avatar, got, c.want)
			}
		})
	}
}

// TestLocalAvatarFilenameCustomPaths 自定义 staticPath / avatarDir 时前缀随之变化
func TestLocalAvatarFilenameCustomPaths(t *testing.T) {
	old := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		StaticPath: "/assets/",
		AvatarDir:  "imgs/",
	}}
	t.Cleanup(func() { conf.Conf = old })

	if got := localAvatarFilename("/assets/imgs/7.png"); got != "7.png" {
		t.Fatalf("自定义路径解析失败: %q", got)
	}
	if got := localAvatarFilename("/static/uploads/avatars/7.png"); got != "" {
		t.Fatalf("非当前前缀不应识别: %q", got)
	}
}
