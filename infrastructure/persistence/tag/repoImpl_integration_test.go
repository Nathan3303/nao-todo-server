//go:build integration

// 集成测试：需要真实 MySQL（docker 一次性容器）+ 本地 Redis
//
//	启动方式：
//	docker run -d --name nao-todo-test-mysql \
//	  -e MYSQL_ROOT_PASSWORD=dev_password -e MYSQL_DATABASE=nao_todo_test \
//	  -p 3307:3306 mysql:8.4.9
//
//	运行：go test -tags integration ./infrastructure/persistence/tag/...
//	可用 NAO_TEST_MYSQL_DSN / NAO_TEST_REDIS_ADDR 覆盖连接串
//	（默认 root:dev_password@tcp(127.0.0.1:3307)/nao_todo_test 与 127.0.0.1:6379）
package tag

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"
)

var testDB *gorm.DB
var testRepo repositories.TagRepository

// TestMain 连接测试库并迁移表结构
func TestMain(m *testing.M) {
	dsn := os.Getenv("NAO_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:dev_password@tcp(127.0.0.1:3307)/nao_todo_test?parseTime=true&loc=UTC&charset=utf8mb4"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("连接测试数据库失败（需先启动测试 MySQL）: %v\n", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(5)

	if err := db.AutoMigrate(&models.Tag{}, &models.TagPreference{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	addr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	rds := redis.NewClient(&redis.Options{Addr: addr})
	testDB = db
	testRepo = NewTagRepo(db, cache.NewCache(rds))
	os.Exit(m.Run())
}

// cleanTags 每用例清理表数据
func cleanTags(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM tag_preferences",
		"DELETE FROM tags",
	} {
		if err := testDB.Exec(stmt).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// newTagVO 构造最小 CreateTag（直接设公开字段）
func newTagVO(id int64, name string, created, updated time.Time) *valueobjects.CreateTag {
	return &valueobjects.CreateTag{
		Id:        id,
		CreatedAt: created,
		UpdatedAt: updated,
		Name:      name,
		Color:     "#FFFFFF",
	}
}

// TestListSyncTombstoneDeletedAt 增量拉取返回墓碑且 deletedAt 非空（pull 修复回归）
func TestListSyncTombstoneDeletedAt(t *testing.T) {
	cleanTags(t)
	ctx := context.Background()
	const userID = 3001
	now := time.Now().UTC().Truncate(time.Millisecond)
	deletedAt := now.Add(-time.Hour)

	// 已删标签（墓碑）：deleted_at 有值
	if err := testDB.WithContext(ctx).Create(&models.Tag{
		ModelBase: models.ModelBase{
			ID:        9101,
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-time.Hour),
			DeletedAt: gorm.DeletedAt{Time: deletedAt, Valid: true},
		},
		UserId: userID,
		Name:   "已删标签",
		Color:  "#FF0000",
	}).Error; err != nil {
		t.Fatalf("插入墓碑: %v", err)
	}
	// 未删标签
	if err := testDB.WithContext(ctx).Create(&models.Tag{
		ModelBase: models.ModelBase{
			ID:        9102,
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now,
		},
		UserId: userID,
		Name:   "正常标签",
		Color:  "#00FF00",
	}).Error; err != nil {
		t.Fatalf("插入正常标签: %v", err)
	}

	list, err := testRepo.ListSync(ctx, userID, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	var tombstone, normal *entities.Tag
	for _, e := range list {
		switch e.Id {
		case 9101:
			tombstone = e
		case 9102:
			normal = e
		}
	}
	if tombstone == nil || normal == nil {
		t.Fatalf("墓碑与正常记录都应被拉取到, got %d 条", len(list))
	}
	if got := tombstone.DeletedAt.ToString(time.RFC3339); got == "" {
		t.Fatal("墓碑 DeletedAt.ToString() = \"\", want 非空删除时间")
	}
	if _, ok := tombstone.DeletedAt.Value(); !ok {
		t.Fatal("墓碑 DeletedAt 无效, want 有效时间")
	}
	if got := normal.DeletedAt.ToString(time.RFC3339); got != "" {
		t.Fatalf("正常记录 DeletedAt.ToString() = %q, want \"\"", got)
	}
}

// TestUpsertTombstoneDeletedAt 推送携带 deletedAt 不复活，未携带则复活
func TestUpsertTombstoneDeletedAt(t *testing.T) {
	cleanTags(t)
	ctx := context.Background()
	const userID = 3002
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 1. 先创建正常标签
	vo := newTagVO(9201, "标签A", now, now)
	if _, _, err := testRepo.Upsert(ctx, userID, vo); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 2. 推送本地墓碑（携带 DeletedAt，updatedAt 更新）→ 不复活，deleted_at 保留
	tombVO := newTagVO(9201, "标签A", now, now.Add(time.Second))
	tombVO.DeletedAt = types.NewNullableTimeByTime(now.Add(30 * time.Second))
	entity, created, err := testRepo.Upsert(ctx, userID, tombVO)
	if err != nil {
		t.Fatalf("墓碑 Upsert: %v", err)
	}
	if created {
		t.Fatal("墓碑 Upsert 应返回 created=false")
	}
	if got := entity.DeletedAt.ToString(time.RFC3339); got == "" {
		t.Fatal("墓碑 Upsert 后 DeletedAt 为空, want 保留删除时间（不应复活）")
	}
	var m models.Tag
	if err := testDB.Unscoped().First(&m, "id = ? AND user_id = ?", 9201, userID).Error; err != nil {
		t.Fatalf("查询: %v", err)
	}
	if !m.DeletedAt.Valid {
		t.Fatal("DB deleted_at 被清空, want 保留墓碑删除时间")
	}

	// 3. 推送正常记录（不携带 DeletedAt，updatedAt 更新）→ 复活
	reviveVO := newTagVO(9201, "标签A", now, now.Add(2*time.Second))
	entity2, _, err := testRepo.Upsert(ctx, userID, reviveVO)
	if err != nil {
		t.Fatalf("复活 Upsert: %v", err)
	}
	if got := entity2.DeletedAt.ToString(time.RFC3339); got != "" {
		t.Fatalf("复活 Upsert 后 DeletedAt = %q, want 空（已复活）", got)
	}
	// 注意：gorm 复用已填充结构体时不会刷新 DeletedAt 字段，故每次查询用新变量
	var mRevive models.Tag
	if err := testDB.Unscoped().First(&mRevive, "id = ? AND user_id = ?", 9201, userID).Error; err != nil {
		t.Fatalf("查询: %v", err)
	}
	if mRevive.DeletedAt.Valid {
		t.Fatal("复活 Upsert 后 DB deleted_at 仍有效, want NULL")
	}
}

// TestUpsertCreateWithDeletedAt 新 id 携带 deletedAt 创建即墓碑
func TestUpsertCreateWithDeletedAt(t *testing.T) {
	cleanTags(t)
	ctx := context.Background()
	const userID = 3003
	now := time.Now().UTC().Truncate(time.Millisecond)

	vo := newTagVO(9301, "创建即删", now, now)
	vo.DeletedAt = types.NewNullableTimeByTime(now.Add(time.Minute))
	entity, created, err := testRepo.Upsert(ctx, userID, vo)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if !created {
		t.Fatal("新 id 应返回 created=true")
	}
	if got := entity.DeletedAt.ToString(time.RFC3339); got == "" {
		t.Fatal("创建即墓碑: DeletedAt 为空, want 删除时间")
	}
	var m models.Tag
	if err := testDB.Unscoped().First(&m, "id = ? AND user_id = ?", 9301, userID).Error; err != nil {
		t.Fatalf("查询: %v", err)
	}
	if !m.DeletedAt.Valid {
		t.Fatal("创建即墓碑: DB deleted_at 无效, want 有值")
	}
}
