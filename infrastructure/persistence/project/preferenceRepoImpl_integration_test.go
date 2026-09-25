//go:build integration

// TASK-26 / T130 契约集成：普通清单偏好 REST（/projects/:projectId/preference）语义不变，
// 按 (user_id, project_id) upsert，updated_at 由服务端写入（按行 LWW 判据）。
//
// 覆盖 qa 用例 CT-02（按行回传）/ CT-09（移动端共用 REST 语义不变）。
package project

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB

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
	if err := db.AutoMigrate(&models.Project{}, &models.ProjectPreference{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	models.InitSnowflake(1)
	testDB = db
	os.Exit(m.Run())
}

// projectPreferenceTestUser 本用例专用用户 ID（避开其它包的测试用户区间）
const projectPreferenceTestUser = 3010

// cleanProjectPreferences 清理本用例用户数据（含满足外键的清单行）
func cleanProjectPreferences(t *testing.T) {
	t.Helper()
	// 仅清理本用例用户，避免与其它包的集成测试（并行执行时共享测试库）互删数据
	for _, stmt := range []string{
		"DELETE FROM project_preferences WHERE user_id = ?",
		"DELETE FROM projects WHERE user_id = ?",
	} {
		if err := testDB.Exec(stmt, projectPreferenceTestUser).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// TestProjectPreference_UpsertByUserAndProjectWithServerUpdatedAt CT-02 / CT-09：
// 同一 (user, project) 按行 upsert（单行），不同 project 独立成行，updated_at 服务端写入且递增。
func TestProjectPreference_UpsertByUserAndProjectWithServerUpdatedAt(t *testing.T) {
	cleanProjectPreferences(t)
	// 用例结束清理：避免遗留孤儿 project_id 行导致其它包 AutoMigrate 建外键失败
	t.Cleanup(func() { cleanProjectPreferences(t) })
	repo := NewProjectPreferenceRepo(testDB)
	ctx := context.Background()
	const userId = projectPreferenceTestUser
	const projectId = 88001

	// 满足 project_preferences.project_id → projects.id 外键
	for _, pid := range []int64{projectId, projectId + 1} {
		if err := testDB.Create(&models.Project{
			ModelBase: models.ModelBase{ID: pid},
			UserId:    userId,
			Name:      "测试清单",
		}).Error; err != nil {
			t.Fatalf("创建测试清单 %d: %v", pid, err)
		}
	}

	first := &valueobjects.SaveProjectPreference{
		ViewType:   types.ViewTypeTable,
		GetOptions: `{"sort":"createdAt"}`,
		Columns:    `["name","state"]`,
	}
	if err := repo.Save(ctx, userId, projectId, first); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Get(ctx, userId, projectId)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got.ViewType) != string(types.ViewTypeTable) ||
		got.GetOptions != first.GetOptions || got.Columns != first.Columns {
		t.Fatalf("首次保存回读不一致: %+v", got)
	}
	if got.UpdatedAt.IsZero() {
		t.Fatal("updated_at 应由服务端写入")
	}

	time.Sleep(5 * time.Millisecond)
	second := &valueobjects.SaveProjectPreference{
		ViewType:   types.ViewTypeKanban,
		GetOptions: `{"sort":"updatedAt"}`,
		Columns:    `["name"]`,
	}
	if err := repo.Save(ctx, userId, projectId, second); err != nil {
		t.Fatalf("Save(二次): %v", err)
	}
	got2, err := repo.Get(ctx, userId, projectId)
	if err != nil {
		t.Fatalf("Get(二次): %v", err)
	}
	if string(got2.ViewType) != string(types.ViewTypeKanban) ||
		got2.GetOptions != second.GetOptions || got2.Columns != second.Columns {
		t.Fatalf("二次保存回读不一致: %+v", got2)
	}
	if !got2.UpdatedAt.After(got.UpdatedAt) {
		t.Fatalf("按行 LWW 需要 updated_at 递增: %v -> %v", got.UpdatedAt, got2.UpdatedAt)
	}

	var count int64
	if err := testDB.Model(&models.ProjectPreference{}).
		Where("user_id = ?", userId).Count(&count).Error; err != nil {
		t.Fatalf("统计: %v", err)
	}
	if count != 1 {
		t.Fatalf("同一 (user, project) 应 upsert 为单行，got %d", count)
	}

	// 不同 project 独立成行
	if err := repo.Save(ctx, userId, projectId+1, first); err != nil {
		t.Fatalf("Save(另一 project): %v", err)
	}
	if err := testDB.Model(&models.ProjectPreference{}).
		Where("user_id = ?", userId).Count(&count).Error; err != nil {
		t.Fatalf("统计(另一 project): %v", err)
	}
	if count != 2 {
		t.Fatalf("不同 project 应独立成行，got %d", count)
	}
}
