//go:build integration

// SYNC-DEF-01 复现/回归：push 空串可空时间字段应清空为 NULL
package task

import (
	"context"
	"database/sql"
	"testing"
	"time"

	apptask "naotodoserver/application/task"
	"naotodoserver/application/task/dto"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"
)

// TestUpsertClearNullableTime_EmptyString 覆盖分支显式空串（Valid=true,IsNull=true）
// ⇒ start_at/end_at/remind_at 落库为 NULL（逐个字段扫描）
func TestUpsertClearNullableTime_EmptyString(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1901
	now := time.Now().UTC().Truncate(time.Millisecond)
	startAt := now.Add(2 * time.Hour)
	endAt := now.Add(3 * time.Hour)
	remindAt := now.Add(1 * time.Hour)

	vo := newTaskVO(9101, "任务", now, now)
	vo.StartAt = types.NewNullableTimeByTime(startAt)
	vo.EndAt = types.NewNullableTimeByTime(endAt)
	vo.RemindAt = types.NewNullableTimeByTime(remindAt)
	if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}

	// 客户端清空：显式空串 → NewNullableTimeByTimeStr("") = {Valid:true, IsNull:true}
	vo2 := newTaskVO(9101, "任务", now, now.Add(2*time.Second))
	vo2.StartAt = types.NewNullableTimeByTimeStr("")
	vo2.EndAt = types.NewNullableTimeByTimeStr("")
	vo2.RemindAt = types.NewNullableTimeByTimeStr("")
	if _, _, err := repo.Upsert(ctx, userID, vo2); err != nil {
		t.Fatalf("清空 Upsert: %v", err)
	}

	var m models.Task
	if err := testDB.First(&m, "id = ?", 9101).Error; err != nil {
		t.Fatalf("回读失败: %v", err)
	}
	if m.StartAt != (sql.NullTime{}) {
		t.Errorf("start_at 期望 NULL, got Valid=%v Time=%v", m.StartAt.Valid, m.StartAt.Time)
	}
	if m.EndAt != (sql.NullTime{}) {
		t.Errorf("end_at 期望 NULL, got Valid=%v Time=%v", m.EndAt.Valid, m.EndAt.Time)
	}
	if m.RemindAt != (sql.NullTime{}) {
		t.Errorf("remind_at 期望 NULL, got Valid=%v Time=%v", m.RemindAt.Valid, m.RemindAt.Time)
	}
}

// TestSyncPush_ClearStartAt_EndToEnd 端到端复现（应用层请求转换 + 仓储 Upsert）：
// 初次推送携带 startAt，二次推送携带 startAt:"" 清空且 endAt 仍有效，
// 期望 DB start_at 为 NULL（修复前：被 FillStartAt 复活为 now）。
func TestSyncPush_ClearStartAt_EndToEnd(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1902
	now := time.Now().UTC().Truncate(time.Second)
	idStr := "9102"
	createdStr := now.Format(time.RFC3339)
	startStr := now.Add(2 * time.Hour).Format(time.RFC3339)
	endStr := now.Add(3 * time.Hour).Format(time.RFC3339)
	remindStr := now.Add(1 * time.Hour).Format(time.RFC3339)

	push := func(updated time.Time, startAt, endAt, remindAt *string) {
		t.Helper()
		req := &dto.CreateTaskReq{
			Id:             &idStr,
			CreatedAt:      &createdStr,
			UpdatedAt:      func() *string { s := updated.Format(time.RFC3339); return &s }(),
			Name:           "任务",
			State:          "pending",
			Priority:       "medium",
			StartAt:        startAt,
			EndAt:          endAt,
			RemindAt:       remindAt,
			ProjectId:      "",
		}
		vo, err := apptask.CreateTaskReqToValueObject(userID, req)
		if err != nil {
			t.Fatalf("CreateTaskReqToValueObject: %v", err)
		}
		if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}

	// 初次推送：携带有效时间
	push(now, &startStr, &endStr, &remindStr)

	// 二次推送：显式空串清空 startAt / remindAt，endAt 保持有效
	empty := ""
	push(now.Add(2*time.Second), &empty, &endStr, &empty)

	var m models.Task
	if err := testDB.First(&m, "id = ?", 9102).Error; err != nil {
		t.Fatalf("回读失败: %v", err)
	}
	if m.StartAt.Valid {
		t.Fatalf("start_at 期望 NULL（显式空串清空），got %v（被 fallback 复活？）", m.StartAt.Time)
	}
	if m.RemindAt.Valid {
		t.Fatalf("remind_at 期望 NULL，got %v", m.RemindAt.Time)
	}
	if !m.EndAt.Valid {
		t.Fatal("end_at 应保持有效（本次未清空）")
	}

	// pull 路径（实体）同样返回空 startAt
	items, err := repo.ListSync(ctx, userID, time.Time{}, 0, 10)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	found := false
	for _, it := range items {
		if it.Id == 9102 {
			found = true
			if v, ok := it.StartAt.Value(); ok {
				t.Fatalf("pull 返回 startAt 应为空，got %v", v)
			}
		}
	}
	if !found {
		t.Fatal("ListSync 未返回 9102")
	}
}

// TestUpdateRepo_EmptyStringClearsNullableTimes 同类字段扫描（PATCH 路径）：
// UpdateTask 请求空串（显式清空）对所有可空时间字段（start/end/remind/archived/starMark/givenUp/completed）
// 均应写 NULL（与 push 路径同型验证）。
func TestUpdateRepo_EmptyStringClearsNullableTimes(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1903
	now := time.Now().UTC().Truncate(time.Second)

	vo := newTaskVO(9103, "任务", now, now)
	vo.StartAt = types.NewNullableTimeByTime(now.Add(2 * time.Hour))
	vo.EndAt = types.NewNullableTimeByTime(now.Add(3 * time.Hour))
	vo.RemindAt = types.NewNullableTimeByTime(now.Add(time.Hour))
	// DEF-SYNC-06：归档/星标/放弃时间已进入 Create 模型，经 Upsert 置位
	vo.ArchivedAt = types.NewNullableTimeByTime(now)
	vo.StarMarkAt = types.NewNullableTimeByTime(now)
	vo.GivenUpAt = types.NewNullableTimeByTime(now)
	if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
		t.Fatalf("初始 Upsert: %v", err)
	}
	// completed_at 仍为纯服务端派生列（不在 Create 模型），直接置位模拟已有值
	if err := testDB.Model(&models.Task{}).
		Where("id = ? AND user_id = ?", 9103, userID).
		Updates(map[string]any{"completed_at": now}).Error; err != nil {
		t.Fatalf("置位完成时间: %v", err)
	}

	empty := ""
	up, err := valueobjects.NewUpdateTask(
		userID, nil, nil, nil, nil, nil,
		&empty, &empty, nil, nil,
		&empty, &empty, &empty, &empty,
		&empty, nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("NewUpdateTask: %v", err)
	}
	if _, err := repo.Update(ctx, userID, 9103, up); err != nil {
		t.Fatalf("Update: %v", err)
	}

	var m models.Task
	if err := testDB.First(&m, "id = ?", 9103).Error; err != nil {
		t.Fatalf("回读失败: %v", err)
	}
	checks := map[string]sql.NullTime{
		"start_at": m.StartAt, "end_at": m.EndAt, "remind_at": m.RemindAt,
		"archived_at": m.ArchivedAt, "star_mark_at": m.StarMarkAt,
		"given_up_at": m.GivenUpAt, "completed_at": m.CompletedAt,
	}
	for col, v := range checks {
		if v.Valid {
			t.Errorf("%s 期望 NULL, got %v", col, v.Time)
		}
	}
}

// TestNewCreateTask_FillStartAtDoesNotResurrectCleared 显式空串 startAt + 有 endAt
// 经 NewCreateTask 后仍应为空（清空不被 FillStartAt 复活为 now）
func TestNewCreateTask_FillStartAtDoesNotResurrectCleared(t *testing.T) {
	empty := ""
	endAt := time.Now().Add(time.Hour).Format(time.RFC3339)
	vo, err := valueobjects.NewCreateTask(
		types.TaskID(0), "任务", "", entities.TaskStatePending, entities.TaskPriorityMedium,
		&empty, &endAt, types.ProjectID(0), nil,
		nil, nil, nil,
		nil, 0, "", 0,
	)
	if err != nil {
		t.Fatalf("NewCreateTask: %v", err)
	}
	if !vo.StartAt.IsSetToNull() {
		t.Fatalf("显式空串 startAt 应保持清空语义，got Valid=%v IsNull=%v Time=%v",
			vo.StartAt.Valid, vo.StartAt.IsNull, vo.StartAt.Time)
	}
}
