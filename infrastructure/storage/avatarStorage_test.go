package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"naotodoserver/conf"
)

const hex32 = "0123456789abcdef0123456789abcdef"

// withTempAvatarConf 设置临时上传目录
func withTempAvatarConf(t *testing.T) string {
	t.Helper()
	uploadDir := t.TempDir()
	old := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		UploadDir:  uploadDir,
		AvatarDir:  "avatars",
		StaticPath: "/static/uploads",
	}}
	t.Cleanup(func() { conf.Conf = old })
	return filepath.Join(uploadDir, "avatars")
}

// TestIsSafeAvatarFilename 文件名白名单（SEC-07）：
// 旧 {userId}.ext 与新 {userId}-{32hex}.ext 均接受；穿越/大小写/长度不符一律拒绝。
func TestIsSafeAvatarFilename(t *testing.T) {
	accept := []string{
		"1.png",
		"12345678901234567890.jpeg",
		"42-" + hex32 + ".png",
		"42-" + hex32 + ".jpg",
	}
	for _, name := range accept {
		if !isSafeAvatarFilename(name) {
			t.Fatalf("%q 应被接受", name)
		}
	}

	reject := []string{
		"",
		".png",
		"1.PNG",
		"1.gif",
		"1.png.exe",
		"../1.png",
		"1/2.png",
		`1\2.png`,
		"42-" + strings.ToUpper(hex32) + ".png", // 大写 hex 不在白名单
		"42-" + hex32[:31] + ".png",             // 31 位
		"42-" + hex32 + "a.png",                 // 33 位
		"42-.png",
		"-" + hex32 + ".png",
		"123456789012345678901.png", // userId 超过 20 位
	}
	for _, name := range reject {
		if isSafeAvatarFilename(name) {
			t.Fatalf("%q 应被拒绝", name)
		}
	}
}

// TestSaveOpenDeleteRoundTrip 新随机名落盘后可读、可删；重复删除幂等
func TestSaveOpenDeleteRoundTrip(t *testing.T) {
	avatarDir := withTempAvatarConf(t)
	store := NewAvatarStorage()
	ctx := context.Background()
	name := "42-" + hex32 + ".png"

	gotURL, err := store.Save(ctx, name, strings.NewReader("avatar-bytes"))
	if err != nil {
		t.Fatalf("保存失败: %v", err)
	}
	if want := "/static/uploads/avatars/" + name; gotURL != want {
		t.Fatalf("相对路径 = %q, want %q", gotURL, want)
	}

	f, err := store.Open(ctx, name)
	if err != nil {
		t.Fatalf("打开失败: %v", err)
	}
	data, err := io.ReadAll(f)
	_ = f.Close()
	if err != nil || string(data) != "avatar-bytes" {
		t.Fatalf("读取内容 = %q / err=%v, want avatar-bytes", data, err)
	}

	if err := store.Delete(ctx, name); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if err := store.Delete(ctx, name); err != nil {
		t.Fatalf("重复删除应幂等: %v", err)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, name)); !os.IsNotExist(err) {
		t.Fatalf("文件应已删除, stat err=%v", err)
	}
}

// TestOpenRejectsUnsafeFilename 非法文件名不得打开（路径穿越防护回归）
func TestOpenRejectsUnsafeFilename(t *testing.T) {
	withTempAvatarConf(t)
	store := NewAvatarStorage()
	for _, name := range []string{"../1.png", "1/2.png", "1.gif", "../../etc/passwd"} {
		if _, err := store.Open(context.Background(), name); err == nil {
			t.Fatalf("%q 不应被打开", name)
		}
	}
}

// TestDeleteRejectsUnsafeFilename 删除同样做白名单校验（SEC-07 新增防护）
func TestDeleteRejectsUnsafeFilename(t *testing.T) {
	withTempAvatarConf(t)
	store := NewAvatarStorage()
	if err := store.Delete(context.Background(), "../1.png"); err == nil {
		t.Fatal("非法文件名不应进入删除路径")
	}
}
