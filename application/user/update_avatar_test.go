package user_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/storage"
)

// readAvatarFixture 读取真实图片样本（与 domain 层共用同一份 testdata）
// @param name 样本文件名
// @return 文件字节
func readAvatarFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(
		"..", "..", "domain", "identity", "valueobjects", "testdata", "avatar", name,
	))
	if err != nil {
		t.Fatalf("读取样本 %s 失败: %v", name, err)
	}
	return data
}

// fakeAvatarUserRepo 只实现头像路径用到的两个方法
type fakeAvatarUserRepo struct {
	repositories.User
	avatar    string
	findErr   error
	updateErr error
}

func (f *fakeAvatarUserRepo) FindById(
	_ context.Context, id domaintypes.UserID,
) (*entities.User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return &entities.User{EntityBase: domaintypes.EntityBase{Id: int64(id)}, Avatar: f.avatar}, nil
}

func (f *fakeAvatarUserRepo) UpdateAvatar(
	_ context.Context, _ domaintypes.UserID, avatarURL string,
) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.avatar = avatarURL
	return nil
}

// fakeCommentApp 只实现评论资料同步（头像上传后会调用）
type fakeCommentApp struct {
	taskApp.TaskCommentApp
	synced int
}

func (f *fakeCommentApp) SyncTaskCommentUserProfile(
	_ context.Context, _ int64, _, _ string,
) error {
	f.synced++
	return nil
}

// newAvatarTestApp 构造「真实本地存储 + 假仓储」的用户应用，返回头像目录
func newAvatarTestApp(
	t *testing.T, userRepo *fakeAvatarUserRepo, commentApp *fakeCommentApp,
) (userApp.UserApp, string) {
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

	app := userApp.NewUserApp(userRepo, nil, commentApp, storage.NewAvatarStorage())
	return app, filepath.Join(uploadDir, "avatars")
}

// TestUpdateAvatarByFileUsesRandomFilename 新上传返回随机名、可读，且把相对路径写入用户表
func TestUpdateAvatarByFileUsesRandomFilename(t *testing.T) {
	userRepo := &fakeAvatarUserRepo{}
	commentApp := &fakeCommentApp{}
	app, avatarDir := newAvatarTestApp(t, userRepo, commentApp)
	content := string(readAvatarFixture(t, "sample.png"))

	out, err := app.UpdateAvatarByFile(
		context.Background(), 42, strings.NewReader(content), "me.png", int64(len(content)),
	)
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	pattern := regexp.MustCompile(`^/static/uploads/avatars/42-[0-9a-f]{32}\.png$`)
	if !pattern.MatchString(out.AvatarURL) {
		t.Fatalf("返回路径 %q 不是不可枚举文件名", out.AvatarURL)
	}
	name := filepath.Base(out.AvatarURL)
	data, err := os.ReadFile(filepath.Join(avatarDir, name))
	if err != nil {
		t.Fatalf("文件未落盘: %v", err)
	}
	if string(data) != content {
		t.Fatalf("文件内容 = %q, want %q", data, content)
	}
	if userRepo.avatar != out.AvatarURL {
		t.Fatalf("用户表头像 = %q, want %q", userRepo.avatar, out.AvatarURL)
	}
	if commentApp.synced != 1 {
		t.Fatalf("评论头像同步 = %d 次, want 1", commentApp.synced)
	}

	// 再次上传得到不同文件名（不会互相覆盖）
	out2, err := app.UpdateAvatarByFile(
		context.Background(), 42, strings.NewReader(content), "me.png", int64(len(content)),
	)
	if err != nil {
		t.Fatalf("二次上传失败: %v", err)
	}
	if out2.AvatarURL == out.AvatarURL {
		t.Fatal("两次上传应生成不同文件名")
	}
}

// TestUpdateAvatarByFileRemovesPreviousLocalFile 上传成功后清理上一张本地头像，
// 避免随机名导致磁盘无界增长；旧 {userId}.ext 在上传前仍可读
func TestUpdateAvatarByFileRemovesPreviousLocalFile(t *testing.T) {
	userRepo := &fakeAvatarUserRepo{avatar: "/static/uploads/avatars/42.png"}
	app, avatarDir := newAvatarTestApp(t, userRepo, &fakeCommentApp{})

	legacyPath := filepath.Join(avatarDir, "42.png")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte("legacy"), 0o600); err != nil {
		t.Fatalf("写入旧头像失败: %v", err)
	}
	// 上传前旧文件名仍可读（路由白名单仍接受 {userId}.ext）
	f, err := os.Open(legacyPath)
	if err != nil {
		t.Fatalf("旧头像应可读: %v", err)
	}
	_ = f.Close()

	fixture := readAvatarFixture(t, "sample.png")
	out, err := app.UpdateAvatarByFile(
		context.Background(), 42, bytes.NewReader(fixture), "me.png", int64(len(fixture)),
	)
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("旧头像文件应被清理，stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, filepath.Base(out.AvatarURL))); err != nil {
		t.Fatalf("新头像文件应存在: %v", err)
	}
}

// TestUpdateAvatarByFileExternalPreviousAvatar 旧头像是外部 URL ⇒ 不做本地删除
func TestUpdateAvatarByFileExternalPreviousAvatar(t *testing.T) {
	userRepo := &fakeAvatarUserRepo{avatar: "https://cdn.example.com/x.png"}
	app, avatarDir := newAvatarTestApp(t, userRepo, &fakeCommentApp{})

	fixture := readAvatarFixture(t, "sample.png")
	out, err := app.UpdateAvatarByFile(
		context.Background(), 42, bytes.NewReader(fixture), "me.png", int64(len(fixture)),
	)
	if err != nil {
		t.Fatalf("上传失败: %v", err)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, filepath.Base(out.AvatarURL))); err != nil {
		t.Fatalf("新头像文件应存在: %v", err)
	}
}

// TestUpdateAvatarByFileFindFailureStillUploads 查询旧头像失败不阻断上传（仅跳过清理）
func TestUpdateAvatarByFileFindFailureStillUploads(t *testing.T) {
	userRepo := &fakeAvatarUserRepo{findErr: io.ErrUnexpectedEOF}
	app, _ := newAvatarTestApp(t, userRepo, &fakeCommentApp{})

	fixture := readAvatarFixture(t, "sample.png")
	if _, err := app.UpdateAvatarByFile(
		context.Background(), 42, bytes.NewReader(fixture), "me.png", int64(len(fixture)),
	); err != nil {
		t.Fatalf("查询旧头像失败不应阻断上传: %v", err)
	}
}

// TestUpdateAvatarByFileValidationRegression 既有校验回归：类型白名单 / 大小上限
func TestUpdateAvatarByFileValidationRegression(t *testing.T) {
	userRepo := &fakeAvatarUserRepo{}
	app, avatarDir := newAvatarTestApp(t, userRepo, &fakeCommentApp{})

	if _, err := app.UpdateAvatarByFile(
		context.Background(), 42, strings.NewReader("gif"), "me.gif", 3,
	); err == nil || !strings.Contains(err.Error(), "头像格式不受支持") {
		t.Fatalf("非图片扩展名应被拒，实际: %v", err)
	}

	tooLarge := int64(2<<20) + 1
	if _, err := app.UpdateAvatarByFile(
		context.Background(), 42, strings.NewReader("x"), "me.png", tooLarge,
	); err == nil || !strings.Contains(err.Error(), "文件大小") {
		t.Fatalf("超限大小应被拒，实际: %v", err)
	}

	// 校验失败不得落盘、不得改用户表
	if entries, err := os.ReadDir(avatarDir); err == nil && len(entries) != 0 {
		t.Fatalf("校验失败不应产生文件，实际 %d 个", len(entries))
	}
	if userRepo.avatar != "" {
		t.Fatalf("校验失败不应更新头像，实际 %q", userRepo.avatar)
	}
}
