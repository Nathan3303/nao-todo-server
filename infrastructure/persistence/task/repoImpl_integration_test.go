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
	// 初始化雪花 ID 节点：服务端生成 id 的路径（如新建检查项）依赖 BeforeCreate 触发
	models.InitSnowflake(1)
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

// TestUpsertTombstoneWithDeletedAt 推送携带 deletedAt（本地墓碑）不复活；未携带则复活
func TestUpsertTombstoneWithDeletedAt(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1006
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 1. 先创建正常任务
	if _, _, err := repo.Upsert(ctx, userID, newTaskVO(9006, "任务", now, now)); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 2. 推送本地墓碑（携带 DeletedAt，updatedAt 更新）→ 不复活，deleted_at 保留
	tombVO := newTaskVO(9006, "任务", now, now.Add(time.Hour))
	tombVO.DeletedAt = types.NewNullableTimeByTime(now.Add(30 * time.Minute))
	entity, created, err := repo.Upsert(ctx, userID, tombVO)
	if err != nil {
		t.Fatalf("墓碑 Upsert: %v", err)
	}
	if created {
		t.Fatal("墓碑 Upsert 应返回 created=false")
	}
	if got := entity.DeletedAt.ToString(time.RFC3339); got == "" {
		t.Fatal("墓碑 Upsert 后 DeletedAt 为空, want 保留删除时间（不应复活）")
	}
	var tomb models.Task
	if err := testDB.Unscoped().First(&tomb, "id = ? AND user_id = ?", 9006, userID).Error; err != nil {
		t.Fatalf("查询: %v", err)
	}
	if !tomb.DeletedAt.Valid {
		t.Fatal("DB deleted_at 被清空, want 保留墓碑删除时间")
	}

	// 3. 推送正常记录（不携带 DeletedAt，updatedAt 更新）→ 复活
	reviveVO := newTaskVO(9006, "任务", now, now.Add(2*time.Hour))
	entity2, _, err := repo.Upsert(ctx, userID, reviveVO)
	if err != nil {
		t.Fatalf("复活 Upsert: %v", err)
	}
	if got := entity2.DeletedAt.ToString(time.RFC3339); got != "" {
		t.Fatalf("复活 Upsert 后 DeletedAt = %q, want 空（已复活）", got)
	}
	// 注意：gorm 复用已填充结构体时不会刷新 DeletedAt 字段，故每次查询用新变量
	var revived models.Task
	if err := testDB.Unscoped().First(&revived, "id = ? AND user_id = ?", 9006, userID).Error; err != nil {
		t.Fatalf("查询: %v", err)
	}
	if revived.DeletedAt.Valid {
		t.Fatal("复活 Upsert 后 DB deleted_at 仍有效, want NULL")
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

// newCheckItemVO 构造最小 CreateTaskCheckItem（绕过校验，直接设公开字段）
func newCheckItemVO(userId int64, id int64, taskId int64, name string, isDone bool, sortId uint16, created, updated time.Time) *valueobjects.CreateTaskCheckItem {
	return &valueobjects.CreateTaskCheckItem{
		UserId:    userId,
		Id:        id,
		CreatedAt: created,
		UpdatedAt: updated,
		TaskId:    taskId,
		Name:      name,
		IsDone:    isDone,
		SortId:    sortId,
	}
}

// TestUpsertCheckItemOverridesIsDoneAndSortId 前端 push 携带 isDone/sortId 覆盖时正确落库
// 回归：此前 Create 链路无 IsDone 字段，push 更新的完成状态被静默丢弃
func TestUpsertCheckItemOverridesIsDoneAndSortId(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1003
	const taskID = 9003
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 首次创建：未完成、sortId 未提供（0，由领域层生成，repo 层原样落库）
	first, created1, err := repo.UpsertCheckItem(ctx, userID, newCheckItemVO(userID, 0, taskID, "How are you ?", false, 0, now, now))
	if err != nil {
		t.Fatalf("首次 UpsertCheckItem: %v", err)
	}
	if !created1 {
		t.Fatal("首次 UpsertCheckItem 应返回 created=true")
	}
	if first.IsDone {
		t.Fatal("首次创建 isDone 应为 false")
	}

	// 客户端 push 更新：同 id 携带 isDone=true、sortId=263（模拟前端请求体字段）
	second, created2, err := repo.UpsertCheckItem(ctx, userID, newCheckItemVO(userID, first.Id, taskID, "How are you ?", true, 263, now, now.Add(2*time.Second)))
	if err != nil {
		t.Fatalf("覆盖 UpsertCheckItem: %v", err)
	}
	if created2 {
		t.Fatal("同 id 覆盖应返回 created=false")
	}
	if !second.IsDone {
		t.Fatal("覆盖后 isDone 应为 true")
	}
	if second.SortId != 263 {
		t.Fatalf("覆盖后 sortId 应为 263，实际 %d", second.SortId)
	}

	// 校验数据库落库
	var m models.TaskCheckItem
	if err := testDB.WithContext(ctx).Where("id = ?", first.Id).First(&m).Error; err != nil {
		t.Fatalf("查询落库记录: %v", err)
	}
	if !m.IsDone {
		t.Fatal("数据库 is_done 应为 true")
	}
	if m.SortId != 263 {
		t.Fatalf("数据库 sort_id 应为 263，实际 %d", m.SortId)
	}
}

// TestRemoveTagFromTasks 标签删除时级联清理任务引用：
// 仅移除精确匹配的 tagId；其他标签与任务字段（如 Name）不被触碰；updated_at 被推进
func TestRemoveTagFromTasks(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1004
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 任务 A：含目标 tagId=1 与相似 tagId=11（子串回归：不应误删）
	taskA, _, err := repo.Upsert(ctx, userID, &valueobjects.CreateTask{
		Id:        9101,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "任务A",
		State:     entities.TaskStatePending,
		Priority:  entities.TaskPriorityMedium,
		Tags:      []string{"1", "11", "2"},
	})
	if err != nil {
		t.Fatalf("创建任务A: %v", err)
	}
	// 任务 B：不含目标 tagId，不应被触碰
	if _, _, err := repo.Upsert(ctx, userID, &valueobjects.CreateTask{
		Id:        9102,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      "任务B",
		State:     entities.TaskStatePending,
		Priority:  entities.TaskPriorityMedium,
		Tags:      []string{"3"},
	}); err != nil {
		t.Fatalf("创建任务B: %v", err)
	}

	if err := repo.RemoveTagFromTasks(ctx, userID, 1); err != nil {
		t.Fatalf("RemoveTagFromTasks: %v", err)
	}

	var mA, mB models.Task
	if err := testDB.WithContext(ctx).First(&mA, "id = ?", taskA.Id).Error; err != nil {
		t.Fatalf("查询任务A: %v", err)
	}
	if err := testDB.WithContext(ctx).First(&mB, "id = ?", 9102).Error; err != nil {
		t.Fatalf("查询任务B: %v", err)
	}

	// 精确移除 tagId=1，保留 11（子串不误删）与 2
	wantTags := []string{"11", "2"}
	if len(mA.Tags) != len(wantTags) {
		t.Fatalf("任务A tags = %v, want %v", mA.Tags, wantTags)
	}
	for i, id := range wantTags {
		if mA.Tags[i] != id {
			t.Fatalf("任务A tags = %v, want %v", mA.Tags, wantTags)
		}
	}
	// 其他字段未被覆盖
	if mA.Name != "任务A" {
		t.Fatalf("任务A Name 被意外覆盖: %q", mA.Name)
	}
	// 任务B 不受影响
	if len(mB.Tags) != 1 || mB.Tags[0] != "3" {
		t.Fatalf("任务B tags = %v, want [3]", mB.Tags)
	}
	// updated_at 被推进（清理事件可被增量同步发现）
	if !mA.UpdatedAt.After(now) && !mA.UpdatedAt.Equal(now.Add(time.Millisecond)) {
		t.Fatalf("任务A updated_at 应被推进, got %v (base %v)", mA.UpdatedAt, now)
	}
}

// TestListCheckItemsSyncTombstone 检查项增量拉取包含软删墓碑
func TestListCheckItemsSyncTombstone(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1005
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 正常检查项
	if _, _, err := repo.UpsertCheckItem(ctx, userID, newCheckItemVO(userID, 9301, 9001, "正常项", false, 1, now, now)); err != nil {
		t.Fatalf("创建正常检查项: %v", err)
	}
	// 墓碑检查项（软删，updated_at 更晚保证排序在后）
	if err := testDB.WithContext(ctx).Unscoped().Create(&models.TaskCheckItem{
		ModelBase: models.ModelBase{
			ID:        9302,
			CreatedAt: now,
			UpdatedAt: now.Add(time.Second),
			DeletedAt: gorm.DeletedAt{Time: now.Add(time.Second), Valid: true},
		},
		UserId: userID,
		TaskId: 9001,
		Name:   "已删项",
		SortId: 2,
	}).Error; err != nil {
		t.Fatalf("创建墓碑检查项: %v", err)
	}

	list, err := repo.ListCheckItemsSync(ctx, userID, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListCheckItemsSync: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("应拉取到 2 条（含墓碑）, got %d", len(list))
	}
	var tombstone, normal *entities.TaskCheckItem
	for _, e := range list {
		switch e.Id {
		case 9302:
			tombstone = e
		case 9301:
			normal = e
		}
	}
	if tombstone == nil || normal == nil {
		t.Fatal("正常检查项与墓碑都应被拉取到")
	}
	if _, ok := tombstone.DeletedAt.Value(); !ok {
		t.Fatal("墓碑 DeletedAt 应有值")
	}
	if _, ok := normal.DeletedAt.Value(); ok {
		t.Fatal("正常检查项 DeletedAt 应为空")
	}
}

// TestListCommentsSyncTombstone 评论增量拉取包含软删墓碑
func TestListCommentsSyncTombstone(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1006
	now := time.Now().UTC().Truncate(time.Millisecond)

	if err := testDB.WithContext(ctx).Unscoped().Create(&models.TaskComment{
		ModelBase: models.ModelBase{
			ID:        9401,
			CreatedAt: now,
			UpdatedAt: now,
		},
		UserId:   userID,
		TaskId:   9001,
		Content:  "正常评论",
		Nickname: "用户",
		Avatar:   "",
	}).Error; err != nil {
		t.Fatalf("创建正常评论: %v", err)
	}
	if err := testDB.WithContext(ctx).Unscoped().Create(&models.TaskComment{
		ModelBase: models.ModelBase{
			ID:        9402,
			CreatedAt: now,
			UpdatedAt: now.Add(time.Second),
			DeletedAt: gorm.DeletedAt{Time: now.Add(time.Second), Valid: true},
		},
		UserId:   userID,
		TaskId:   9001,
		Content:  "已删评论",
		Nickname: "用户",
		Avatar:   "",
	}).Error; err != nil {
		t.Fatalf("创建墓碑评论: %v", err)
	}

	list, err := repo.ListCommentsSync(ctx, userID, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListCommentsSync: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("应拉取到 2 条（含墓碑）, got %d", len(list))
	}
	var tombstone, normal *entities.TaskComment
	for _, e := range list {
		switch e.Id {
		case 9402:
			tombstone = e
		case 9401:
			normal = e
		}
	}
	if tombstone == nil || normal == nil {
		t.Fatal("正常评论与墓碑都应被拉取到")
	}
	if _, ok := tombstone.DeletedAt.Value(); !ok {
		t.Fatal("墓碑 DeletedAt 应有值")
	}
	if _, ok := normal.DeletedAt.Value(); ok {
		t.Fatal("正常评论 DeletedAt 应为空")
	}
}

// TestListCheckItemsSyncKeyset 检查项增量拉取 keyset 游标分页不重不漏
func TestListCheckItemsSyncKeyset(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1007
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 同秒 5 条：updated_at 相同，靠 id 二级推进
	for i, id := range []int64{9501, 9502, 9503, 9504, 9505} {
		if _, _, err := repo.UpsertCheckItem(ctx, userID, newCheckItemVO(userID, id, 9001, "项", false, uint16(i+1), now, now)); err != nil {
			t.Fatalf("创建检查项 %d: %v", id, err)
		}
	}

	var got []int64
	cursor, cursorID := time.Time{}, int64(0)
	for {
		page, err := repo.ListCheckItemsSync(ctx, userID, cursor, cursorID, 2)
		if err != nil {
			t.Fatalf("ListCheckItemsSync: %v", err)
		}
		if len(page) == 0 {
			break
		}
		for _, e := range page {
			got = append(got, e.Id)
		}
		last := page[len(page)-1]
		cursor, cursorID = last.UpdatedAt, last.Id
		if len(got) > 5 {
			t.Fatalf("keyset 推进出现重复: %v", got)
		}
	}
	if len(got) != 5 {
		t.Fatalf("keyset 推进遗漏: got %d, want 5 (%v)", len(got), got)
	}
	want := []int64{9501, 9502, 9503, 9504, 9505}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("顺序不一致: got %v, want %v", got, want)
		}
	}
}

// TestCreateStartAtRoundtrip DEF-SYNC-04 端到端：客户端显式 startAt 经真实 MySQL 往返后
// 秒级瞬时不变（写路径秒级截断，不被服务端 now 覆盖，也不被 DATETIME 小数秒四舍五入漂移）
func TestCreateStartAtRoundtrip(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1010

	// .900 毫秒：若不截断，MySQL 会四舍五入到下一秒（41→42），秒级往返即被破坏
	startAt, _ := time.Parse(time.RFC3339Nano, "2026-09-11T16:28:41.900+08:00")
	endAt, _ := time.Parse(time.RFC3339Nano, "2026-09-11T20:00:00.123+08:00")
	now := time.Now().UTC().Truncate(time.Millisecond)

	vo := newTaskVO(9601, "往返任务", now, now)
	vo.StartAt = types.NewNullableTimeByTime(startAt)
	vo.EndAt = types.NewNullableTimeByTime(endAt)

	created, err := repo.Create(ctx, userID, vo)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	gotStart, ok := created.StartAt.Value()
	if !ok || gotStart.Unix() != startAt.Truncate(time.Second).Unix() {
		t.Fatalf("创建响应 startAt 秒级漂移: got %v ok=%v, want %d", gotStart, ok, startAt.Truncate(time.Second).Unix())
	}

	// 拉回（等价 pull 路径）同样秒级不变
	fetched, err := repo.GetById(ctx, userID, 9601, false)
	if err != nil {
		t.Fatalf("GetById: %v", err)
	}
	gotStart, ok = fetched.StartAt.Value()
	if !ok || gotStart.Unix() != startAt.Truncate(time.Second).Unix() {
		t.Fatalf("往返后 startAt 秒级漂移: got %v ok=%v, want %d", gotStart, ok, startAt.Truncate(time.Second).Unix())
	}
	gotEnd, ok := fetched.EndAt.Value()
	if !ok || gotEnd.Unix() != endAt.Truncate(time.Second).Unix() {
		t.Fatalf("往返后 endAt 秒级漂移: got %v ok=%v, want %d", gotEnd, ok, endAt.Truncate(time.Second).Unix())
	}
}
