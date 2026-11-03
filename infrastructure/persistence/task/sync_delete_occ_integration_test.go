//go:build integration

// T642 / A4 集成测试（真实 MySQL）：删除纳入 OCC + 回传真实 updated_at。
//
// 覆盖：
//   - 删除回执的 updated_at **逐字等于**库中值（R-20：控制器不再用循环前的 now）
//   - base 不匹配 ⇒ 不写库、**不产生新墓碑**（查库证明 updated_at / deleted_at 均未变）
//   - 老客户端（无 base）⇒ 行为与改前一致（无条件软删）
//   - 已是墓碑的同向重复删除 ⇒ 幂等 applied，不推进 updated_at
//
// 依赖 repoImpl_integration_test.go 的 TestMain 与 count_events_integration_test.go 的装配。
package task

import (
	"context"
	"strings"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	taskDto "naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"

	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroDto "naotodoserver/application/pomodoro/dto"
	projectDto "naotodoserver/application/project/dto"
	tagApp "naotodoserver/application/tag"
	tagDto "naotodoserver/application/tag/dto"
	pomodoroService "naotodoserver/domain/pomodoro/service"
	tagService "naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	pomodoroRepo "naotodoserver/infrastructure/persistence/pomodoro"
	tagRepo "naotodoserver/infrastructure/persistence/tag"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm/clause"
)

// deleteOCCNewTask 创建一条任务并返回 (id, 库中 updated_at)
func deleteOCCNewTask(t *testing.T, s *fullStack, idStr string) int64 {
	t.Helper()
	req := taskReq("删除OCC", 0, 0, &idStr)
	req.CreatedAt = deleteOCCStrPtr(time.Now().UTC().Format(time.RFC3339))
	req.UpdatedAt = req.CreatedAt
	id, err := idutil.ParseID(idStr)
	if err != nil {
		t.Fatalf("解析 id %q: %v", idStr, err)
	}
	if _, _, err := s.taskApp.CreateTask(context.Background(), testUserID, req); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	return id
}

func deleteOCCStrPtr(s string) *string { return &s }

// deleteOCCDBRow 读库中任务的 updated_at / deleted_at（逐字比对用）
func deleteOCCDBRow(t *testing.T, taskID int64) (updatedAt string, deleted bool) {
	t.Helper()
	var m models.Task
	if err := testDB.Unscoped().Where("id = ? AND user_id = ?", taskID, testUserID).First(&m).Error; err != nil {
		t.Fatalf("读任务 %d: %v", taskID, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli), m.DeletedAt.Valid
}

// TestDeleteTask_OCC_ReturnsRealTimestamp base 命中 ⇒ 软删，回执时间戳 == 库中值（逐字）
func TestDeleteTask_OCC_ReturnsRealTimestamp(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	taskID := deleteOCCNewTask(t, s, "9701")

	beforeUpdatedAt, beforeDeleted := deleteOCCDBRow(t, taskID)
	if beforeDeleted {
		t.Fatal("新建任务不应是墓碑")
	}

	row, decision, err := s.taskApp.DeleteTask(
		context.Background(), testUserID, idutil.FormatID(taskID), mustParseMilli(t, beforeUpdatedAt),
	)
	if err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if decision != domaintypes.DeleteApply {
		t.Fatalf("decision = %v, want DeleteApply", decision)
	}
	if row == nil {
		t.Fatal("base 命中的删除必须回传库中当前行")
	}

	afterUpdatedAt, afterDeleted := deleteOCCDBRow(t, taskID)
	if !afterDeleted {
		t.Fatal("base 命中后库中 deleted_at 必须非空（软删）")
	}
	if row.UpdatedAt != afterUpdatedAt {
		t.Fatalf("回执 updatedAt = %q，库中 = %q（必须逐字相等，否则客户端写回 base 后永久 stale）",
			row.UpdatedAt, afterUpdatedAt)
	}
	if row.DeletedAt == "" {
		t.Fatal("回执必须携带 deletedAt（墓碑可见）")
	}
	if afterUpdatedAt == beforeUpdatedAt {
		t.Fatal("软删必须推进 updated_at")
	}
}

// TestDeleteTask_OCC_BaseMismatchKeepsRow 不匹配 ⇒ stale，不写库、不产生新墓碑
func TestDeleteTask_OCC_BaseMismatchKeepsRow(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	taskID := deleteOCCNewTask(t, s, "9702")

	beforeUpdatedAt, _ := deleteOCCDBRow(t, taskID)
	// 客户端持有的版本比库中早 1ms（模拟他人已改）
	staleBase := mustParseMilli(t, beforeUpdatedAt).Add(-time.Millisecond)

	row, decision, err := s.taskApp.DeleteTask(
		context.Background(), testUserID, idutil.FormatID(taskID), staleBase,
	)
	if err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if decision != domaintypes.DeleteStale {
		t.Fatalf("decision = %v, want DeleteStale", decision)
	}
	if row == nil || row.DeletedAt != "" {
		t.Fatalf("stale 必须回传库中当前行（未删除），实际 = %+v", row)
	}
	if row.UpdatedAt != beforeUpdatedAt {
		t.Fatalf("stale 回传的 updatedAt = %q，库中 = %q", row.UpdatedAt, beforeUpdatedAt)
	}

	afterUpdatedAt, afterDeleted := deleteOCCDBRow(t, taskID)
	if afterDeleted {
		t.Fatal("base 不匹配不得产生新墓碑（deleted_at 必须仍为空）")
	}
	if afterUpdatedAt != beforeUpdatedAt {
		t.Fatalf("base 不匹配不得推进 updated_at：before=%q after=%q", beforeUpdatedAt, afterUpdatedAt)
	}
}

// TestDeleteTask_NoBase_OldClientUnchanged 老客户端（无 base）⇒ 无条件软删，行为与改前一致（R-D）
func TestDeleteTask_NoBase_OldClientUnchanged(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	taskID := deleteOCCNewTask(t, s, "9703")

	row, decision, err := s.taskApp.DeleteTask(
		context.Background(), testUserID, idutil.FormatID(taskID), time.Time{},
	)
	if err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if decision != domaintypes.DeleteApply {
		t.Fatalf("decision = %v, want DeleteApply", decision)
	}
	afterUpdatedAt, afterDeleted := deleteOCCDBRow(t, taskID)
	if !afterDeleted {
		t.Fatal("无 base 的删除必须软删（删除优先，逐字旧行为）")
	}
	if row == nil || row.UpdatedAt != afterUpdatedAt || row.DeletedAt == "" {
		t.Fatalf("回执必须等于库中墓碑行：row=%+v dbUpdatedAt=%q", row, afterUpdatedAt)
	}
}

// TestDeleteTask_RepeatDeleteIdempotent 已是墓碑 ⇒ 幂等 applied，不推进 updated_at（不产生新墓碑）
func TestDeleteTask_RepeatDeleteIdempotent(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	taskID := deleteOCCNewTask(t, s, "9704")

	if _, _, err := s.taskApp.DeleteTask(
		context.Background(), testUserID, idutil.FormatID(taskID), time.Time{},
	); err != nil {
		t.Fatalf("首次删除: %v", err)
	}
	tombstoneUpdatedAt, deleted := deleteOCCDBRow(t, taskID)
	if !deleted {
		t.Fatal("首次删除后应为墓碑")
	}

	row, decision, err := s.taskApp.DeleteTask(
		context.Background(), testUserID, idutil.FormatID(taskID),
		mustParseMilli(t, tombstoneUpdatedAt),
	)
	if err != nil {
		t.Fatalf("重复删除: %v", err)
	}
	if decision != domaintypes.DeleteIdempotent {
		t.Fatalf("decision = %v, want DeleteIdempotent", decision)
	}
	afterUpdatedAt, _ := deleteOCCDBRow(t, taskID)
	if afterUpdatedAt != tombstoneUpdatedAt {
		t.Fatalf("重复删除不得推进 updated_at：before=%q after=%q", tombstoneUpdatedAt, afterUpdatedAt)
	}
	if row == nil || row.UpdatedAt != tombstoneUpdatedAt {
		t.Fatalf("重复删除必须回传库中墓碑行（updatedAt=%q），实际 = %+v", tombstoneUpdatedAt, row)
	}
}

// TestDeleteTaskCheckItemAndComment_OCC 检查项/评论同样回传真实时间戳（表覆盖抽查）
func TestDeleteTaskCheckItemAndComment_OCC(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	taskID := deleteOCCNewTask(t, s, "9705")
	ctx := context.Background()

	itemID := "9801"
	ciRes, _, err := s.taskApp.CreateTaskCheckItem(ctx, testUserID, &taskDto.CreateTaskCheckItemReq{
		TaskId: idutil.FormatID(taskID), Name: "检查项", Id: &itemID,
	})
	if err != nil {
		t.Fatalf("CreateTaskCheckItem: %v", err)
	}
	itemDBUpdatedAt := deleteOCCCheckItemUpdatedAt(t, itemIDInt(t, ciRes.Id))
	ciRow, ciDecision, err := s.taskApp.DeleteTaskCheckItem(
		ctx, testUserID, ciRes.Id, mustParseMilli(t, itemDBUpdatedAt),
	)
	if err != nil {
		t.Fatalf("DeleteTaskCheckItem: %v", err)
	}
	if ciDecision != domaintypes.DeleteApply {
		t.Fatalf("checkItem decision = %v, want DeleteApply", ciDecision)
	}
	if ciRow == nil || ciRow.UpdatedAt != deleteOCCCheckItemUpdatedAt(t, itemIDInt(t, ciRes.Id)) {
		t.Fatalf("检查项回执 updatedAt 必须等于库中值，row=%+v", ciRow)
	}

	// 评论创建需要 users 行（仓储取昵称/头像），补一行测试用户（重复跑幂等）
	if err := testDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.User{
		ModelBase: models.ModelBase{ID: testUserID},
		Account:   "delete-occ-tester",
		Email:     "delete-occ-tester@example.com",
		Password:  strings.Repeat("x", 64),
		Nickname:  "测试用户",
	}).Error; err != nil {
		t.Fatalf("插入测试用户: %v", err)
	}

	commentID := "9802"
	cmRes, _, err := s.taskApp.CreateTaskComment(ctx, testUserID, &taskDto.CreateTaskCommentReq{
		TaskId: idutil.FormatID(taskID), Content: "评论", Id: &commentID,
	})
	if err != nil {
		t.Fatalf("CreateTaskComment: %v", err)
	}
	cmDBUpdatedAt := deleteOCCCommentUpdatedAt(t, itemIDInt(t, cmRes.Id))
	cmRow, cmDecision, err := s.taskApp.DeleteTaskComment(
		ctx, testUserID, cmRes.Id, mustParseMilli(t, cmDBUpdatedAt),
	)
	if err != nil {
		t.Fatalf("DeleteTaskComment: %v", err)
	}
	if cmDecision != domaintypes.DeleteApply {
		t.Fatalf("comment decision = %v, want DeleteApply", cmDecision)
	}
	if cmRow == nil || cmRow.UpdatedAt != deleteOCCCommentUpdatedAt(t, itemIDInt(t, cmRes.Id)) {
		t.Fatalf("评论回执 updatedAt 必须等于库中值，row=%+v", cmRow)
	}
}

// itemIDInt 解析雪花/客户端预置 ID
func itemIDInt(t *testing.T, id string) int64 {
	t.Helper()
	v, err := idutil.ParseID(id)
	if err != nil {
		t.Fatalf("解析 id %q: %v", id, err)
	}
	return v
}

func deleteOCCCheckItemUpdatedAt(t *testing.T, id int64) string {
	t.Helper()
	var m models.TaskCheckItem
	if err := testDB.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
		t.Fatalf("读检查项 %d: %v", id, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli)
}

func deleteOCCCommentUpdatedAt(t *testing.T, id int64) string {
	t.Helper()
	var m models.TaskComment
	if err := testDB.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
		t.Fatalf("读评论 %d: %v", id, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli)
}

// mustParseMilli 解析 RFC3339Milli（与回执口径一致）
func mustParseMilli(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := idutil.ParseBaseUpdatedAt(&s)
	if err != nil {
		t.Fatalf("解析时间 %q: %v", s, err)
	}
	return parsed
}

// TestDeleteProject_OCC_DeactivedAtTombstone 清单的删除标记是 deactived_at（非 GORM deleted_at）：
// base 不匹配 ⇒ stale（不写库、不级联删任务）；base 命中 ⇒ 软删 + 级联，回执时间戳 == 库中值。
func TestDeleteProject_OCC_DeactivedAtTombstone(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	ctx := context.Background()

	projectID := "9901"
	if _, _, err := s.projApp.Create(ctx, testUserID, &projectDto.CreateProjectReq{
		Id: &projectID, Name: "清单",
	}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	projectIDInt := itemIDInt(t, projectID)

	subTaskID := "9902"
	if _, _, err := s.taskApp.CreateTask(ctx, testUserID, taskReq("子任务", projectIDInt, 0, &subTaskID)); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	subTaskIDInt := itemIDInt(t, subTaskID)

	beforeUpdatedAt := deleteOCCProjectUpdatedAt(t, projectIDInt)

	// 1) base 不匹配 ⇒ stale，不写库、不级联
	if _, decision, err := s.projApp.Delete(
		ctx, testUserID, projectID, mustParseMilli(t, beforeUpdatedAt).Add(-time.Millisecond),
	); err != nil {
		t.Fatalf("Delete(stale): %v", err)
	} else if decision != domaintypes.DeleteStale {
		t.Fatalf("decision = %v, want DeleteStale", decision)
	}
	if got := deleteOCCProjectUpdatedAt(t, projectIDInt); got != beforeUpdatedAt {
		t.Fatalf("stale 不得推进清单 updated_at：before=%q after=%q", beforeUpdatedAt, got)
	}
	if _, deleted := deleteOCCDBRow(t, subTaskIDInt); deleted {
		t.Fatal("stale 清单删除不得级联删掉子任务")
	}

	// 2) base 命中 ⇒ 软删（deactived_at）+ 级联 + 回执真实时间戳
	row, decision, err := s.projApp.Delete(ctx, testUserID, projectID, mustParseMilli(t, beforeUpdatedAt))
	if err != nil {
		t.Fatalf("Delete(applied): %v", err)
	}
	if decision != domaintypes.DeleteApply {
		t.Fatalf("decision = %v, want DeleteApply", decision)
	}
	afterUpdatedAt := deleteOCCProjectUpdatedAt(t, projectIDInt)
	if row == nil || row.UpdatedAt != afterUpdatedAt {
		t.Fatalf("回执 updatedAt 必须等于库中值：row=%+v db=%q", row, afterUpdatedAt)
	}
	if row.DeactivedAt == "" {
		t.Fatal("清单软删必须体现在 deactivedAt（客户端据此判定删除）")
	}
	if _, deleted := deleteOCCDBRow(t, subTaskIDInt); !deleted {
		t.Fatal("清单软删必须级联软删下属任务（既有行为不得回退）")
	}
}

// TestDeleteTagAndPomodoro_OCC 标签/常用番茄工作：回执时间戳 == 库中值，base 不匹配 ⇒ 不产生新墓碑
func TestDeleteTagAndPomodoro_OCC(t *testing.T) {
	cleanTasks(t)
	if err := testDB.AutoMigrate(&models.Tag{}, &models.TagPreference{}, &models.Pomodoro{}); err != nil {
		t.Fatalf("迁移标签/番茄表: %v", err)
	}
	for _, stmt := range []string{"DELETE FROM tags", "DELETE FROM tag_preferences", "DELETE FROM pomodoros"} {
		if err := testDB.Exec(stmt).Error; err != nil {
			t.Fatalf("清理 %s: %v", stmt, err)
		}
	}
	// 标签：复用任务仓的装配挂载 tag app（级联清理任务引用需要 taskDomain）
	tagRepoInst := tagRepo.NewTagRepo(testDB, cache.NewCache(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})))
	tagDomainInst := tagService.NewTagDomain(tagRepoInst, tagRepo.NewTagPreferenceRepo(testDB))
	txManager := dbs.NewTxManager(testDB)
	taskRepoInst := NewTaskRepo(testDB)
	taskDomainInst := taskService.NewTaskDomain(taskRepoInst, taskRepoInst)
	tagAppInst := tagApp.NewTagApp(tagDomainInst, tagRepoInst, tagRepo.NewTagPreferenceRepo(testDB), txManager, taskDomainInst)

	ctx := context.Background()
	tagID := "9911"
	if _, _, err := tagAppInst.CreateTag(ctx, testUserID, &tagDto.CreateTagReq{Id: &tagID, Name: "标签", Color: "#FF0000"}); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	tagIDInt := itemIDInt(t, tagID)
	tagUpdatedAt := deleteOCCTagUpdatedAt(t, tagIDInt)

	if _, decision, err := tagAppInst.DeleteTag(
		ctx, testUserID, tagID, mustParseMilli(t, tagUpdatedAt).Add(-time.Millisecond),
	); err != nil {
		t.Fatalf("DeleteTag(stale): %v", err)
	} else if decision != domaintypes.DeleteStale {
		t.Fatalf("tag decision = %v, want DeleteStale", decision)
	}
	if got := deleteOCCTagUpdatedAt(t, tagIDInt); got != tagUpdatedAt {
		t.Fatalf("stale 不得推进标签 updated_at：before=%q after=%q", tagUpdatedAt, got)
	}

	tagRow, tagDecision, err := tagAppInst.DeleteTag(ctx, testUserID, tagID, mustParseMilli(t, tagUpdatedAt))
	if err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	if tagDecision != domaintypes.DeleteApply {
		t.Fatalf("tag decision = %v, want DeleteApply", tagDecision)
	}
	if dbTagUpdatedAt := deleteOCCTagUpdatedAt(t, tagIDInt); tagRow == nil || tagRow.UpdatedAt != dbTagUpdatedAt {
		t.Fatalf("标签回执 updatedAt 必须等于库中值，row=%+v db=%q", tagRow, dbTagUpdatedAt)
	}

	// 常用番茄工作
	pomodoroRepoInst := pomodoroRepo.NewPomodoroRepo(testDB)
	pomodoroDomainInst := pomodoroService.NewPomodoroDomain(pomodoroRepo.NewPomodoroRecordRepo(testDB), pomodoroRepoInst)
	pomodoroAppInst := pomodoroApp.NewPomodoroApp(pomodoroDomainInst, pomodoroRepo.NewPomodoroRecordRepo(testDB), pomodoroRepoInst)

	pomodoroID := "9921"
	if _, _, err := pomodoroAppInst.CreatePomodoro(ctx, testUserID, &pomodoroDto.CreatePomodoroReq{
		Id: &pomodoroID, Name: "专注", Duration: 1500,
	}); err != nil {
		t.Fatalf("CreatePomodoro: %v", err)
	}
	pomodoroIDInt := itemIDInt(t, pomodoroID)
	pomodoroUpdatedAt := deleteOCCPomodoroUpdatedAt(t, pomodoroIDInt)

	if _, decision, err := pomodoroAppInst.DeletePomodoro(
		ctx, testUserID, pomodoroID, mustParseMilli(t, pomodoroUpdatedAt).Add(time.Millisecond),
	); err != nil {
		t.Fatalf("DeletePomodoro(stale): %v", err)
	} else if decision != domaintypes.DeleteStale {
		t.Fatalf("pomodoro decision = %v, want DeleteStale", decision)
	}
	if got := deleteOCCPomodoroUpdatedAt(t, pomodoroIDInt); got != pomodoroUpdatedAt {
		t.Fatalf("stale 不得推进番茄 updated_at：before=%q after=%q", pomodoroUpdatedAt, got)
	}

	pomodoroRow, pomodoroDecision, err := pomodoroAppInst.DeletePomodoro(
		ctx, testUserID, pomodoroID, mustParseMilli(t, pomodoroUpdatedAt),
	)
	if err != nil {
		t.Fatalf("DeletePomodoro: %v", err)
	}
	if pomodoroDecision != domaintypes.DeleteApply {
		t.Fatalf("pomodoro decision = %v, want DeleteApply", pomodoroDecision)
	}
	if pomodoroRow == nil || pomodoroRow.UpdatedAt != deleteOCCPomodoroUpdatedAt(t, pomodoroIDInt) {
		t.Fatalf("番茄回执 updatedAt 必须等于库中值，row=%+v", pomodoroRow)
	}
}

func deleteOCCProjectUpdatedAt(t *testing.T, id int64) string {
	t.Helper()
	var m models.Project
	if err := testDB.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
		t.Fatalf("读清单 %d: %v", id, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli)
}

func deleteOCCTagUpdatedAt(t *testing.T, id int64) string {
	t.Helper()
	var m models.Tag
	if err := testDB.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
		t.Fatalf("读标签 %d: %v", id, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli)
}

func deleteOCCPomodoroUpdatedAt(t *testing.T, id int64) string {
	t.Helper()
	var m models.Pomodoro
	if err := testDB.Unscoped().Where("id = ?", id).First(&m).Error; err != nil {
		t.Fatalf("读常用番茄 %d: %v", id, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli)
}
