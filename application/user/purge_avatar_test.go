package user_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	taskApp "naotodoserver/application/task"
	"naotodoserver/application/user"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/infrastructure/storage"

	"github.com/sirupsen/logrus"
)

// fakePurgeUserRepo 只实现硬删清理所需方法
type fakePurgeUserRepo struct {
	repositories.User
	purged []*entities.User
	err    error
	calls  int
	days   []int8
}

func (f *fakePurgeUserRepo) DeleteDeactivatedUsers(
	_ context.Context, dayOffset int8,
) ([]*entities.User, error) {
	f.calls++
	f.days = append(f.days, dayOffset)
	return f.purged, f.err
}

// fakePurgeStorage 记录删除调用，可注入错误
type fakePurgeStorage struct {
	deleted []string
	err     error
}

func (f *fakePurgeStorage) Save(_ context.Context, _ string, _ io.Reader) (string, error) {
	return "", nil
}

func (f *fakePurgeStorage) Open(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, os.ErrNotExist
}

func (f *fakePurgeStorage) Delete(_ context.Context, filename string) error {
	if f.err != nil {
		return f.err
	}
	f.deleted = append(f.deleted, filename)
	return nil
}

// withUploadsConfig 头像路径工具依赖 conf，构造测试用配置
func withUploadsConfig(t *testing.T, uploadDir string) {
	t.Helper()
	old := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		UploadDir:  uploadDir,
		AvatarDir:  "avatars",
		StaticPath: "/static/uploads",
	}}
	t.Cleanup(func() { conf.Conf = old })
}

// capturePurgeLogs 捕获结构化日志
func capturePurgeLogs(t *testing.T) *bytesBuffer {
	t.Helper()
	old := logging.Logger
	logger := logrus.New()
	buf := &bytesBuffer{}
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logging.Logger = logger
	t.Cleanup(func() { logging.Logger = old })
	return buf
}

// bytesBuffer 最小日志缓冲（避免测试额外依赖）
type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) String() string { return string(b.data) }

// newPurgeApp 构造"硬删清理"用户应用
func newPurgeApp(repo *fakePurgeUserRepo, avatarStorage user.AvatarStorage) user.UserApp {
	var commentApp taskApp.TaskCommentApp = nil
	return user.NewUserApp(repo, nil, commentApp, avatarStorage)
}

// TestPurgeAvatarOnHardDelete ① 本地头像（新旧命名）⇒ 删除；② ⑤ 外部 URL / 非本用户 ⇒ 跳过
func TestPurgeAvatarOnHardDelete(t *testing.T) {
	withUploadsConfig(t, t.TempDir())
	hex32 := strings.Repeat("a", 32)
	cases := map[string]struct {
		userId     int64
		avatar     string
		wantDelete string
	}{
		"旧命名":   {42, "/static/uploads/avatars/42.png", "42.png"},
		"新命名":   {43, "/static/uploads/avatars/43-" + hex32 + ".webp", "43-" + hex32 + ".webp"},
		"外部URL": {44, "https://cdn.example.com/a.png", ""},
		"他人文件":  {45, "/static/uploads/avatars/999-" + hex32 + ".png", ""},
		"非法名字":  {46, "/static/uploads/avatars/../etc/passwd", ""},
		"空头像":   {47, "", ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			storage := &fakePurgeStorage{}
			repo := &fakePurgeUserRepo{purged: []*entities.User{{
				EntityBase: domaintypes.EntityBase{Id: tc.userId},
				Avatar:     tc.avatar,
			}}}
			app := newPurgeApp(repo, storage)

			if err := app.DeleteDeactivatedUsers(context.Background(), 7); err != nil {
				t.Fatalf("清理失败: %v", err)
			}
			if repo.calls != 1 || len(repo.days) != 1 || repo.days[0] != 7 {
				t.Fatalf("仓储调用 = %d 次 / 天数 %v, want 1 次 / [7]", repo.calls, repo.days)
			}
			if tc.wantDelete == "" {
				if len(storage.deleted) != 0 {
					t.Fatalf("不应删除任何文件，实际 %v", storage.deleted)
				}
				return
			}
			if len(storage.deleted) != 1 || storage.deleted[0] != tc.wantDelete {
				t.Fatalf("删除列表 = %v, want [%s]", storage.deleted, tc.wantDelete)
			}
		})
	}
}

// TestPurgeAvatarFailureDoesNotBlock ① 文件删除失败 ⇒ Warn 且清理任务仍成功（日志不含路径全量）
func TestPurgeAvatarFailureDoesNotBlock(t *testing.T) {
	withUploadsConfig(t, t.TempDir())
	logs := capturePurgeLogs(t)
	storage := &fakePurgeStorage{err: errors.New("permission denied: /tmp/secret/avatars/42.png")}
	repo := &fakePurgeUserRepo{purged: []*entities.User{{
		EntityBase: domaintypes.EntityBase{Id: 42},
		Avatar:     "/static/uploads/avatars/42.png",
	}}}
	app := newPurgeApp(repo, storage)

	if err := app.DeleteDeactivatedUsers(context.Background(), 7); err != nil {
		t.Fatalf("文件删除失败不应阻断清理任务: %v", err)
	}
	logText := logs.String()
	if !strings.Contains(logText, "avatar.purge_failed") || !strings.Contains(logText, "42") {
		t.Fatalf("应记录含 user_id 的 Warn 日志: %s", logText)
	}
	if strings.Contains(logText, "/tmp/secret") {
		t.Fatalf("日志不得包含路径全量: %s", logText)
	}
}

// TestPurgeAvatarRepoErrorPropagates 仓储失败 ⇒ 返回错误（不静默）
func TestPurgeAvatarRepoErrorPropagates(t *testing.T) {
	withUploadsConfig(t, t.TempDir())
	repo := &fakePurgeUserRepo{err: errors.New("db down")}
	app := newPurgeApp(repo, &fakePurgeStorage{})
	if err := app.DeleteDeactivatedUsers(context.Background(), 7); err == nil {
		t.Fatal("仓储失败必须返回错误")
	}
}

// TestPurgeAvatarNilStorageAndEmptyBatch 无待删用户 / 未注入存储 ⇒ 不 panic、不报错
func TestPurgeAvatarNilStorageAndEmptyBatch(t *testing.T) {
	withUploadsConfig(t, t.TempDir())
	app := newPurgeApp(&fakePurgeUserRepo{}, nil)
	if err := app.DeleteDeactivatedUsers(context.Background(), 7); err != nil {
		t.Fatalf("空批次不应报错: %v", err)
	}

	app = newPurgeApp(&fakePurgeUserRepo{purged: []*entities.User{{
		EntityBase: domaintypes.EntityBase{Id: 42},
		Avatar:     "/static/uploads/avatars/42.png",
	}}}, nil)
	if err := app.DeleteDeactivatedUsers(context.Background(), 7); err != nil {
		t.Fatalf("未注入存储时不应报错: %v", err)
	}
}

// TestPurgeAvatarRealStorageOwnership ④ 真实文件系统：只删本人文件，他人文件与不存在文件都安全
func TestPurgeAvatarRealStorageOwnership(t *testing.T) {
	uploadDir := t.TempDir()
	withUploadsConfig(t, uploadDir)
	avatarDir := filepath.Join(uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		t.Fatalf("建目录失败: %v", err)
	}
	hex32 := strings.Repeat("b", 32)
	mine := "42-" + hex32 + ".png"
	others := "43-" + hex32 + ".png"
	for _, name := range []string{mine, others} {
		if err := os.WriteFile(filepath.Join(avatarDir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("写文件失败: %v", err)
		}
	}

	repo := &fakePurgeUserRepo{purged: []*entities.User{
		{EntityBase: domaintypes.EntityBase{Id: 42}, Avatar: "/static/uploads/avatars/" + mine},
		// 同一批里另一个用户头像文件不存在（幂等/不存在不报错）
		{EntityBase: domaintypes.EntityBase{Id: 44}, Avatar: "/static/uploads/avatars/44-" + hex32 + ".png"},
	}}
	app := newPurgeApp(repo, storage.NewAvatarStorage())

	if err := app.DeleteDeactivatedUsers(context.Background(), 7); err != nil {
		t.Fatalf("清理失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, mine)); !os.IsNotExist(err) {
		t.Fatalf("本人头像应被删除，stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, others)); err != nil {
		t.Fatalf("他人头像不得被删: %v", err)
	}
}
