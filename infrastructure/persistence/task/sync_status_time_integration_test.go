//go:build integration

// DEF-SYNC-06 复现/回归：sync push 状态时间戳（archived_at / star_mark_at / given_up_at）
// 三态落库（首次置值 → 幂等同值 → 二次空串清空 → 回读一致）。
//
// 修复前：桌面端 push 携带 givenUpAt 等字段被 Go 反序列化静默丢弃，且推送仍返回成功（假成功），
// 客户端清队列后下次 pull 被服务端 NULL 覆盖 ⇒「放弃」回退。本用例确保字段真正落库。
package task

import (
	"context"
	"testing"
	"time"

	apptask "naotodoserver/application/task"
	"naotodoserver/application/task/dto"
	"naotodoserver/infrastructure/persistence/models"
)

// TestSyncPush_StatusTimestamps_EndToEnd 端到端（应用层请求转换 + 仓储 Upsert + ListSync 回读）
func TestSyncPush_StatusTimestamps_EndToEnd(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1904
	idStr := "9201"

	now := time.Now().UTC().Truncate(time.Second)
	createdStr := now.Format(time.RFC3339)
	archived := now.Add(-1 * time.Hour)
	star := now.Add(-2 * time.Hour)
	givenUp := now.Add(-30 * time.Minute)

	strPtr := func(tm time.Time) *string { v := tm.Format(time.RFC3339); return &v }

	// push 模拟 POST /sync/push 的任务条目 → app 转换 → 仓储 upsert
	push := func(updated time.Time, archivedAt, starMarkAt, givenUpAt *string) {
		t.Helper()
		updatedStr := updated.Format(time.RFC3339)
		req := &dto.CreateTaskReq{
			Id:         &idStr,
			CreatedAt:  &createdStr,
			UpdatedAt:  &updatedStr,
			Name:       "任务",
			State:      "pending",
			Priority:   "medium",
			ProjectId:  "",
			ArchivedAt: archivedAt,
			StarMarkAt: starMarkAt,
			GivenUpAt:  givenUpAt,
		}
		vo, err := apptask.CreateTaskReqToValueObject(userID, req)
		if err != nil {
			t.Fatalf("CreateTaskReqToValueObject: %v", err)
		}
		if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
	}

	readModel := func() models.Task {
		t.Helper()
		var m models.Task
		if err := testDB.First(&m, "id = ?", 9201).Error; err != nil {
			t.Fatalf("回读失败: %v", err)
		}
		return m
	}

	// assertListSync 校验 pull 路径回读的三字段有效性
	assertListSync := func(wantValid bool) {
		t.Helper()
		items, err := repo.ListSync(ctx, userID, time.Time{}, 0, 10)
		if err != nil {
			t.Fatalf("ListSync: %v", err)
		}
		var found bool
		for _, it := range items {
			if it.Id != 9201 {
				continue
			}
			found = true
			_, aOK := it.ArchivedAt.Value()
			_, sOK := it.StarMarkAt.Value()
			_, gOK := it.GivenUpAt.Value()
			if aOK != wantValid || sOK != wantValid || gOK != wantValid {
				t.Fatalf("ListSync 状态时间戳有效性 = %v/%v/%v, want %v", aOK, sOK, gOK, wantValid)
			}
		}
		if !found {
			t.Fatal("ListSync 未返回 9201")
		}
	}

	// ① 首次 push 置值 ⇒ DB 实际写入非空（假成功防线）
	push(now, strPtr(archived), strPtr(star), strPtr(givenUp))
	m := readModel()
	if !m.ArchivedAt.Valid || !m.StarMarkAt.Valid || !m.GivenUpAt.Valid {
		t.Fatalf("首次 push 状态时间戳未落库（推送假成功）: archived=%v star=%v givenUp=%v",
			m.ArchivedAt, m.StarMarkAt, m.GivenUpAt)
	}
	if !m.ArchivedAt.Time.Equal(archived) || !m.StarMarkAt.Time.Equal(star) || !m.GivenUpAt.Time.Equal(givenUp) {
		t.Fatalf("落库值与推送值不一致: got %v/%v/%v, want %v/%v/%v",
			m.ArchivedAt.Time, m.StarMarkAt.Time, m.GivenUpAt.Time, archived, star, givenUp)
	}
	assertListSync(true)

	// ② 幂等：同值再次 push（updatedAt 更晚）⇒ 不矛盾、值不变
	push(now.Add(2*time.Second), strPtr(archived), strPtr(star), strPtr(givenUp))
	m = readModel()
	if !m.GivenUpAt.Valid || !m.GivenUpAt.Time.Equal(givenUp) {
		t.Fatalf("幂等重复 push 后 given_up_at 不一致: %v", m.GivenUpAt)
	}

	// ③ 二次 push 空串 ⇒ 三列置 NULL（取消放弃/归档/星标）
	empty := ""
	push(now.Add(4*time.Second), &empty, &empty, &empty)
	m = readModel()
	if m.ArchivedAt.Valid || m.StarMarkAt.Valid || m.GivenUpAt.Valid {
		t.Fatalf("显式空串应清空三列: archived=%v star=%v givenUp=%v",
			m.ArchivedAt, m.StarMarkAt, m.GivenUpAt)
	}
	assertListSync(false)

	// ④ 部分置位（只重新放弃）⇒ 各列独立
	push(now.Add(6*time.Second), nil, nil, strPtr(givenUp))
	m = readModel()
	if !m.GivenUpAt.Valid || !m.GivenUpAt.Time.Equal(givenUp) {
		t.Fatalf("重新置值 given_up_at 失败: %v", m.GivenUpAt)
	}
	if m.ArchivedAt.Valid || m.StarMarkAt.Valid {
		t.Fatalf("缺省字段不应触碰既有 NULL: archived=%v star=%v", m.ArchivedAt, m.StarMarkAt)
	}
}
