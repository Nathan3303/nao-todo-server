//go:build integration

// H1 复现/回归：提醒扫描仅返回「活跃」任务——已完成 / 已归档 / 已放弃 / 已软删
// 的任务不得出现在到期提醒列表，正常 todo / in-progress 任务仍正常返回。
package task

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/types"
)

// dbNow 读取数据库当前时间，作为「到期 / 未到期」的造数基准。
//
// DEF-158：生产侧筛选条件是 SQL `remind_at <= NOW()`，`NOW()` 用的是 **DB 会话时区**
// （本机 MySQL 为 Asia/Shanghai），而入库时间由驱动按 DSN 的 `loc=UTC` 格式化。
// 若测试用 `time.Now().UTC()` 造数，两类时间就错开了时区偏移（本机 8 小时）
// ⇒ 明明是「1 小时后到期」的任务也会被 `NOW()` 判为已到期。
//
// 这里改用 DB 自己的当前时钟（读出 → 原样写回）：驱动读、写两端都用 `loc`，往返是**同一个 wall-clock**，
// 故造出的 due/future 与 `NOW()` 同域。判据语义与生产完全一致（就是「相对数据库当前时间」），
// 断言一个都没有改动、也没有放宽。
func dbNow(t *testing.T) time.Time {
	t.Helper()
	var current sql.NullTime
	if err := testDB.Raw("SELECT NOW()").Scan(&current).Error; err != nil {
		t.Fatalf("读取数据库当前时间: %v", err)
	}
	if !current.Valid {
		t.Fatal("读取数据库当前时间为 NULL")
	}
	return current.Time.Truncate(time.Second)
}

// TestGetDueReminders_ExcludesInactive 到期扫描排除维度扫描：
// state=done / archived_at 非空 / given_up_at 非空 / 未到期 / remind_at NULL / 软删
func TestGetDueReminders_ExcludesInactive(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1910

	// 以数据库时钟为基准造数（DEF-158：不能用 time.Now() 造，否则与 SQL NOW() 错时区）
	now := dbNow(t)
	due := now.Add(-time.Minute) // 已到期
	future := now.Add(time.Hour) // 未到期

	mk := func(id int64, state entities.TaskState, remindAt *time.Time, archived, givenUp bool) {
		t.Helper()
		vo := newTaskVO(id, "任务", now, now)
		vo.State = state
		if remindAt != nil {
			vo.RemindAt = types.NewNullableTimeByTime(*remindAt)
		}
		if archived {
			vo.ArchivedAt = types.NewNullableTimeByTime(now)
		}
		if givenUp {
			vo.GivenUpAt = types.NewNullableTimeByTime(now)
		}
		if _, _, err := repo.Upsert(ctx, userID, vo); err != nil {
			t.Fatalf("Upsert %d: %v", id, err)
		}
	}

	// 应返回：待办 / 进行中，且已到期
	mk(9301, entities.TaskStatePending, &due, false, false)
	mk(9302, entities.TaskStateInProgress, &due, false, false)
	// 应排除
	mk(9303, entities.TaskStateCompleted, &due, false, false)  // 已完成
	mk(9304, entities.TaskStatePending, &due, true, false)     // 已归档
	mk(9305, entities.TaskStatePending, &due, false, true)     // 已放弃
	mk(9306, entities.TaskStatePending, &future, false, false) // 未到期
	mk(9307, entities.TaskStatePending, nil, false, false)     // remind_at 为空
	mk(9308, entities.TaskStatePending, &due, false, false)    // 已软删（下方 Delete）
	mk(9309, entities.TaskStateCompleted, &due, true, true)    // 三重非活跃

	if err := repo.Delete(ctx, userID, 9308); err != nil {
		t.Fatalf("软删 9308: %v", err)
	}

	got, err := repo.GetDueReminders(ctx)
	if err != nil {
		t.Fatalf("GetDueReminders: %v", err)
	}
	gotIDs := make(map[int64]struct{}, len(got))
	for _, task := range got {
		gotIDs[task.Id] = struct{}{}
	}
	for _, want := range []int64{9301, 9302} {
		if _, ok := gotIDs[want]; !ok {
			t.Errorf("活跃任务 %d 应从到期列表返回，但缺失", want)
		}
	}
	for _, excluded := range []int64{9303, 9304, 9305, 9306, 9307, 9308, 9309} {
		if _, ok := gotIDs[excluded]; ok {
			t.Errorf("非活跃/未到期任务 %d 不应出现在到期列表", excluded)
		}
	}
}
