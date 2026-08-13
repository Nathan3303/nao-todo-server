//go:build integration

// 集成测试：需要真实 MySQL（docker 一次性容器）
//
//	启动方式：
//	docker run -d --name nao-todo-test-mysql \
//	  -e MYSQL_ROOT_PASSWORD=dev_password -e MYSQL_DATABASE=nao_todo_test \
//	  -p 3307:3306 mysql:8.4.9
//
//	运行：go test -tags integration ./infrastructure/persistence/task/...
//	可用 NAO_TEST_MYSQL_DSN 覆盖连接串（默认 root:dev_password@tcp(127.0.0.1:3307)）
package task

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"
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
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(5)

	if err := db.AutoMigrate(&models.Task{}, &models.TaskCheckItem{}, &models.TaskComment{}); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	testDB = db
	os.Exit(m.Run())
}

// cleanTasks 每用例清理表数据
func cleanTasks(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM task_comments",
		"DELETE FROM task_check_items",
		"DELETE FROM tasks",
	} {
		if err := testDB.Exec(stmt).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// newTaskVO 构造最小 CreateTask（绕过 NewCreateTask 校验，直接设公开字段）
func newTaskVO(id int64, name string, created, updated time.Time) *valueobjects.CreateTask {
	return &valueobjects.CreateTask{
		Id:        id,
		CreatedAt: created,
		UpdatedAt: updated,
		Name:      name,
		State:     entities.TaskStatePending,
		Priority:  entities.TaskPriorityMedium,
	}
}

// TestUpsertIdempotent 同 id 重复推送不重复创建
func TestUpsertIdempotent(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1001
	now := time.Now().UTC().Truncate(time.Millisecond)

	first, created1, err := repo.Upsert(ctx, userID, newTaskVO(9001, "任务A", now, now))
	if err != nil {
		t.Fatalf("首次 Upsert: %v", err)
	}
	if !created1 {
		t.Fatal("首次 Upsert 应返回 created=true")
	}

	second, created2, err := repo.Upsert(ctx, userID, newTaskVO(9001, "任务A", now, now.Add(2*time.Second)))
	if err != nil {
		t.Fatalf("重复 Upsert: %v", err)
	}
	if created2 {
		t.Fatal("重复 Upsert 应返回 created=false")
	}
	if first.Id != second.Id {
		t.Fatalf("id 不一致: %d != %d", first.Id, second.Id)
	}

	var cnt int64
	testDB.Model(&models.Task{}).Where("id = ? AND user_id = ?", 9001, userID).Count(&cnt)
	if cnt != 1 {
		t.Fatalf("同 id 应只有一行, got %d", cnt)
	}
}

// TestUpsertLWW 旧 updatedAt 不覆盖新数据，新 updatedAt 覆盖
func TestUpsertLWW(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1002
	now := time.Now().UTC().Truncate(time.Millisecond)

	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9002, "原始", now, now)); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 旧 updatedAt（早于库中版本）→ no-op，内容不变
	_, _, err := repo.Upsert(ctx, userID, newTaskVO(9002, "旧数据", now, now.Add(-time.Minute)))
	if err != nil {
		t.Fatalf("旧版本 Upsert: %v", err)
	}
	var m models.Task
	testDB.First(&m, "id = ?", 9002)
	if m.Name != "原始" {
		t.Fatalf("旧 updatedAt 不应覆盖, got name=%q", m.Name)
	}

	// 新 updatedAt → 覆盖
	_, _, err = repo.Upsert(ctx, userID, newTaskVO(9002, "新数据", now, now.Add(time.Minute)))
	if err != nil {
		t.Fatalf("新版本 Upsert: %v", err)
	}
	testDB.First(&m, "id = ?", 9002)
	if m.Name != "新数据" {
		t.Fatalf("新 updatedAt 应覆盖, got name=%q", m.Name)
	}
}

// TestUpsertCreateConflict createdAt 与库中差异过大 → ErrIDConflict（ID 碰撞）
func TestUpsertCreateConflict(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1003
	now := time.Now().UTC().Truncate(time.Millisecond)

	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9003, "实体", now, now)); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 不同实体撞 id：createdAt 相差 > 1 分钟
	other := newTaskVO(9003, "另一实体", now.Add(-2*time.Hour), now)
	_, _, err := repo.Upsert(ctx, userID, other)
	if err == nil {
		t.Fatal("createdAt 冲突应返回 ErrIDConflict")
	}
	if err != domerr.ErrIDConflict {
		t.Fatalf("应返回 ErrIDConflict, got %v", err)
	}
}

// TestUpsertServerTime 覆盖后落库 updated_at 为服务器时间（与客户端时间无关）
func TestUpsertServerTime(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1004
	now := time.Now().UTC().Truncate(time.Millisecond)
	clientFuture := now.Add(48 * time.Hour) // 客户端时钟严重偏快

	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9004, "任务", now, clientFuture)); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 覆盖：客户端 updatedAt 偏快 48h，落库应为服务器 now 而非客户端时间
	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9004, "任务v2", now, clientFuture)); err != nil {
		t.Fatalf("覆盖 Upsert: %v", err)
	}
	var m models.Task
	testDB.First(&m, "id = ?", 9004)
	if m.UpdatedAt.After(time.Now().Add(10 * time.Minute)) {
		t.Fatalf("落库 updated_at 应为服务器时间, got %v（疑似客户端时间）", m.UpdatedAt)
	}
}

// TestUpsertReviveTombstone 软删（墓碑）后同 id Upsert 覆盖应复活（deleted_at 清空）
func TestUpsertReviveTombstone(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1005
	now := time.Now().UTC().Truncate(time.Millisecond)

	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9005, "任务", now, now)); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}
	if err := repo.Delete(ctx, userID, 9005); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	var tomb models.Task
	testDB.Unscoped().First(&tomb, "id = ?", 9005)
	if tomb.DeletedAt.Valid == false {
		t.Fatal("Delete 后应产生软删墓碑")
	}

	// 客户端重试推送同 id（未感知删除）→ 覆盖复活
	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9005, "任务v2", now, now.Add(time.Minute))); err != nil {
		t.Fatalf("复活 Upsert: %v", err)
	}
	var revived models.Task
	testDB.Unscoped().First(&revived, "id = ?", 9005)
	if revived.DeletedAt.Valid {
		t.Fatal("覆盖后 deleted_at 应被清空（墓碑复活）")
	}
	testDB.First(&revived, "id = ?", 9005) // 非 Unscoped 应可见
	if revived.Name != "任务v2" {
		t.Fatalf("复活后名称应为新值, got %q", revived.Name)
	}
}

// TestListSyncKeyset 同秒多记录 keyset 分页推进不重复不遗漏，墓碑在增量中可见
func TestListSyncKeyset(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1006
	ts := time.Now().UTC().Truncate(time.Millisecond)

	// 同一时刻 5 条记录（同 updated_at，id 递增）
	for i := int64(0); i < 5; i++ {
		vo := newTaskVO(9100+i, fmt.Sprintf("任务%d", i), ts, ts)
		vo.SortId = uint16(i)
		if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
			t.Fatalf("Upsert %d: %v", i, err)
		}
	}
	// 其中一条软删（墓碑）
	if err := repo.Delete(ctx, userID, 9103); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// 逐页 keyset 拉取（每页 2 条）
	var got []*entities.Task
	cursor, cursorID := time.Time{}, int64(0)
	for {
		page, err := repo.ListSync(ctx, userID, cursor, cursorID, 2)
		if err != nil {
			t.Fatalf("ListSync: %v", err)
		}
		if len(page) == 0 {
			break
		}
		for _, e := range page {
			for _, prev := range got {
				if prev.Id == e.Id {
					t.Fatalf("keyset 推进重复: id=%d", e.Id)
				}
			}
		}
		got = append(got, page...)
		last := page[len(page)-1]
		cursor, cursorID = last.UpdatedAt, last.Id
		if len(got) > 6 {
			t.Fatal("keyset 推进超出预期")
		}
	}

	if len(got) != 5 {
		t.Fatalf("keyset 应返回 5 条（含墓碑）, got %d", len(got))
	}
	// 墓碑记录应含 deletedAt（客户端据此本地删除）
	for _, e := range got {
		if e.Id == 9103 {
			if _, ok := e.DeletedAt.Value(); !ok {
				t.Fatal("软删任务的墓碑应在增量中可见（deletedAt 非空）")
			}
		}
	}
}

// TestReminderCAS ClearRemindRepeat/UpdateRemindAt 仅当 remind_at 仍为期望值时生效
func TestReminderCAS(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1007
	now := time.Now().UTC().Truncate(time.Millisecond)
	dueAt := time.Now().UTC().Truncate(time.Second).Add(-time.Minute) // 扫描时到期的 remind_at（秒级，与 Format(RFC3339) 一致）
	snoozedAt := now.Add(30 * time.Minute)                            // 用户 Snooze 后的新 remind_at

	vo := newTaskVO(9200, "提醒任务", now, now)
	vo.RemindAt = types.NewNullableTimeByTimeStr(dueAt.Format(time.RFC3339))
	vo.RemindRepeat = 1
	vo.RemindTime = "09:00"
	vo.RemindWeekdays = 5
	if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 场景 A：remind_at 未被改期 → CAS 生效，配置被清空
	changed, err := repo.ClearRemindRepeat(ctx, 9200, dueAt)
	if err != nil {
		t.Fatalf("ClearRemindRepeat: %v", err)
	}
	if !changed {
		t.Fatal("期望值匹配时应返回 changed=true")
	}
	var m models.Task
	testDB.First(&m, "id = ?", 9200)
	if m.RemindAt.Valid || m.RemindRepeat != 0 || m.RemindTime != "" || m.RemindWeekdays != 0 {
		t.Fatalf("CAS 生效后提醒字段应全清: %+v", m)
	}

	// 场景 B：模拟 Snooze（remind_at 已变为新值），再用旧期望值 CAS → 不生效
	testDB.Model(&models.Task{}).Where("id = ?", 9200).
		Update("remind_at", snoozedAt)
	changed, err = repo.ClearRemindRepeat(ctx, 9200, dueAt)
	if err != nil {
		t.Fatalf("ClearRemindRepeat(旧期望): %v", err)
	}
	if changed {
		t.Fatal("remind_at 已变更时 CAS 不应生效")
	}
	testDB.First(&m, "id = ?", 9200)
	if !m.RemindAt.Valid {
		t.Fatal("CAS 不生效时 remind_at 应保留 Snooze 后的值")
	}
	if m.RemindAt.Time.Truncate(time.Second) != snoozedAt.Truncate(time.Second) {
		t.Fatalf("remind_at 应保留 Snooze 值, got %v", m.RemindAt.Time)
	}
}
