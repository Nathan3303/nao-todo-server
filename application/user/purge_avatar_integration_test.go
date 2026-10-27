//go:build integration

// DEF-106 集成回归（真实 MySQL + 真实本地存储）：
// 注销 7 天等待期到期的用户，在每日硬删任务中被物理删除，同时清理其本服务头像文件；
// 未到期用户不受影响，他人头像文件不误删，文件不存在不报错（幂等）。
//
// 运行：go test -tags integration -p 1 -count=1 -run TestPurgeDeactivatedUser -v ./application/user/
package user_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	userApp "naotodoserver/application/user"
	"naotodoserver/conf"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/identity"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/storage"

	"github.com/go-redis/redis/v8"
	sqldriver "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newPurgeTestDB 为硬删清理用例创建独立临时库并迁移 11 张表：
// 避免把 users/user_configs 的 FK 约束带进共享测试库，从而干扰其它包的既有用例。
// @param t 测试上下文
// @return 独立库连接
func newPurgeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("NAO_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:dev_password@tcp(127.0.0.1:3307)/nao_todo_test" +
			"?parseTime=true&loc=UTC&charset=utf8mb4"
	}
	cfg, err := sqldriver.ParseDSN(dsn)
	if err != nil {
		t.Fatalf("解析测试 DSN 失败: %v", err)
	}
	adminCfg := *cfg
	adminCfg.DBName = ""
	admin, err := gorm.Open(gormmysql.Open(adminCfg.FormatDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("连接测试 MySQL 失败: %v", err)
	}
	dbName := fmt.Sprintf("nao_todo_purge_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE DATABASE `" + dbName + "` CHARACTER SET utf8mb4").Error; err != nil {
		t.Fatalf("创建临时库失败: %v", err)
	}
	t.Cleanup(func() {
		_ = admin.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error
		sqlDB, _ := admin.DB()
		_ = sqlDB.Close()
	})

	cfg.DBName = dbName
	db, err := gorm.Open(gormmysql.Open(cfg.FormatDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// 独立库内不需要 FK 约束：清理逻辑按 user_id 显式删除 11 张表
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("连接临时库失败: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	if err := db.AutoMigrate(
		&models.User{}, &models.UserSession{}, &models.UserConfig{},
		&models.Project{}, &models.ProjectPreference{},
		&models.Tag{}, &models.TagPreference{},
		&models.Task{}, &models.TaskCheckItem{}, &models.TaskComment{},
		&models.Pomodoro{}, &models.PomodoroRecord{},
	); err != nil {
		t.Fatalf("迁移临时库失败: %v", err)
	}
	models.InitSnowflake(1)
	return db
}

// seedDeactivatedUser 插入一个已注销用户（deactivedAt 天之前），返回 userId
func seedDeactivatedUser(t *testing.T, db *gorm.DB, email string, daysAgo int) int64 {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte("legacy-pass-1"), 4)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	user := &models.User{
		Account:  email,
		Email:    email,
		Password: string(hashed),
		Nickname: "purge",
	}
	user.DeactivedAt = sql.NullTime{Time: time.Now().AddDate(0, 0, -daysAgo), Valid: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("插入注销用户失败: %v", err)
	}
	return user.ID
}

// TestPurgeDeactivatedUserIntegration 到期用户 ⇒ 硬删 + 头像清理；未到期用户与关联数据回归
func TestPurgeDeactivatedUserIntegration(t *testing.T) {
	ctx := context.Background()
	purgeDB := newPurgeTestDB(t)
	uploadDir := t.TempDir()
	oldConf := conf.Conf
	conf.Conf = &conf.Config{Uploads: &conf.Uploads{
		UploadDir:  uploadDir,
		AvatarDir:  "avatars",
		StaticPath: "/static/uploads",
	}}
	t.Cleanup(func() { conf.Conf = oldConf })

	avatarDir := filepath.Join(uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		t.Fatalf("建头像目录失败: %v", err)
	}

	suffix := time.Now().UnixNano()
	expiredEmail := "purge-expired-" + strconv.FormatInt(suffix, 10) + "@example.test"
	aliveEmail := "purge-alive-" + strconv.FormatInt(suffix, 10) + "@example.test"

	// 到期用户：先建用户拿到 ID，再按「{userId}-{32hex}.png」命名头像文件
	expiredId := seedDeactivatedUser(t, purgeDB, expiredEmail, 8)
	expiredAvatarName := strconv.FormatInt(expiredId, 10) + "-" + strings.Repeat("c", 32) + ".png"
	expiredAvatar := "/static/uploads/avatars/" + expiredAvatarName
	if err := purgeDB.Model(&models.User{}).Where("id = ?", expiredId).
		Update("avatar", expiredAvatar).Error; err != nil {
		t.Fatalf("写头像字段失败: %v", err)
	}
	expiredPath := filepath.Join(avatarDir, expiredAvatarName)
	if err := os.WriteFile(expiredPath, []byte("avatar"), 0o600); err != nil {
		t.Fatalf("写头像失败: %v", err)
	}
	// 该用户的关联数据（验证 11 张表清理仍工作）
	project := &models.Project{
		ModelBase: models.ModelBase{ID: time.Now().UnixNano()},
		UserId:    expiredId,
		Name:      "purge-project",
		SortId:    1000,
	}
	if err := purgeDB.Create(project).Error; err != nil {
		t.Fatalf("插入关联清单失败: %v", err)
	}

	// 未到期用户（2 天）：不得被删
	aliveId := seedDeactivatedUser(t, purgeDB, aliveEmail, 2)
	aliveAvatarName := strconv.FormatInt(aliveId, 10) + "-" + strings.Repeat("d", 32) + ".png"
	if err := purgeDB.Model(&models.User{}).Where("id = ?", aliveId).
		Update("avatar", "/static/uploads/avatars/"+aliveAvatarName).Error; err != nil {
		t.Fatalf("写未到期用户头像字段失败: %v", err)
	}
	alivePath := filepath.Join(avatarDir, aliveAvatarName)
	if err := os.WriteFile(alivePath, []byte("avatar"), 0o600); err != nil {
		t.Fatalf("写未到期用户头像失败: %v", err)
	}

	rds := redis.NewClient(&redis.Options{
		Addr:        redisAddrOrLocal(),
		Password:    os.Getenv("NAO_TEST_REDIS_PASSWORD"),
		DialTimeout: 50 * time.Millisecond,
		MaxRetries:  -1,
	})
	t.Cleanup(func() { _ = rds.Close() })
	app := userApp.NewUserApp(
		identity.NewUserRepo(purgeDB, cache.NewCache(rds)), nil, nil, storage.NewAvatarStorage(),
	)
	if err := app.DeleteDeactivatedUsers(ctx, 7); err != nil {
		t.Fatalf("硬删清理失败: %v", err)
	}

	// ① 到期用户与本服务头像文件均被删除
	var userCount int64
	if err := purgeDB.Model(&models.User{}).Where("id = ?", expiredId).Count(&userCount).Error; err != nil {
		t.Fatalf("回读用户失败: %v", err)
	}
	if userCount != 0 {
		t.Fatalf("到期用户未被硬删，剩余 %d 行", userCount)
	}
	if _, err := os.Stat(expiredPath); !os.IsNotExist(err) {
		t.Fatalf("到期用户头像文件应被删除，stat err=%v", err)
	}

	// ⑤ 清理任务其余行为：关联数据同步删除
	var projectCount int64
	if err := purgeDB.Model(&models.Project{}).Where("id = ?", project.ID).Count(&projectCount).Error; err != nil {
		t.Fatalf("回读清单失败: %v", err)
	}
	if projectCount != 0 {
		t.Fatalf("关联清单未被删除，剩余 %d 行", projectCount)
	}

	// 未到期用户与其头像不受影响
	var aliveCount int64
	if err := purgeDB.Model(&models.User{}).Where("id = ?", aliveId).Count(&aliveCount).Error; err != nil {
		t.Fatalf("回读未到期用户失败: %v", err)
	}
	if aliveCount != 1 {
		t.Fatalf("未到期用户不得被删，剩余 %d 行", aliveCount)
	}
	if _, err := os.Stat(alivePath); err != nil {
		t.Fatalf("未到期用户头像不得被删: %v", err)
	}

	// 幂等：再跑一次无副作用、不报错
	if err := app.DeleteDeactivatedUsers(ctx, 7); err != nil {
		t.Fatalf("重复执行应幂等且不报错: %v", err)
	}

	// 回归：未到期用户仍可正常查询（用户数据完整）
	if _, err := identity.NewUserRepo(purgeDB, cache.NewCache(rds)).
		FindByEmail(ctx, aliveEmail); err != nil {
		t.Fatalf("未到期用户应仍可查询: %v", err)
	}
}

// redisAddrOrLocal 测试用 Redis 地址（连不上也不影响本用例：缓存失效静默降级）
func redisAddrOrLocal() string {
	if addr := os.Getenv("NAO_TEST_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379"
}
