// H1 回归：提醒扫描口径补齐（repo WHERE 过滤）不得影响领域层既有行为——
// 重复提醒触发后按规则改期、一次性提醒触发后自愈清空。
package service

import (
	"context"
	"testing"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/types"
)

// fakeRemindRepo 仅实现 ProcessReminders 路径用到的方法；其余由内嵌 nil 接口占位。
type fakeRemindRepo struct {
	repositories.Task
	due     []*entities.Task
	updated map[int64]string
	cleared map[int64]bool
}

func (f *fakeRemindRepo) GetDueReminders(ctx context.Context) ([]*entities.Task, error) {
	return f.due, nil
}

func (f *fakeRemindRepo) UpdateRemindAt(
	ctx context.Context,
	taskId int64,
	expectedRemindAt time.Time,
	remindAt string,
) (bool, error) {
	f.updated[taskId] = remindAt
	return true, nil
}

func (f *fakeRemindRepo) ClearRemindRepeat(
	ctx context.Context,
	taskId int64,
	expectedRemindAt time.Time,
) (bool, error) {
	f.cleared[taskId] = true
	return true, nil
}

// TestProcessReminders_RescheduleAndSelfHeal 重复提醒改期 / 一次性提醒自愈，
// 且无 end_at 的重复提醒必须正常续期（H4：零值 end_at 不再被误判为已越过终点）。
func TestProcessReminders_RescheduleAndSelfHeal(t *testing.T) {
	due := time.Date(2026, 9, 21, 9, 0, 0, 0, time.UTC) // 周一
	// weekly：次日（周二）匹配（calculateNextWeekly 位：1<<Weekday，周二=4）
	weeklyMask := uint8(1) << uint(due.AddDate(0, 0, 1).Weekday())

	fake := &fakeRemindRepo{
		due: []*entities.Task{
			{ // daily，无 end_at → 续期到次日
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatDaily),
				RemindTime:   "09:00",
				RemindAt:     types.NewNullableTimeByTime(due),
			},
			{ // 一次性 → 自愈清空
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatNone),
				RemindAt:     types.NewNullableTimeByTime(due),
			},
			{ // daily，有 end_at 且下一次越过 → 清空（回归）
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatDaily),
				RemindTime:   "09:00",
				RemindAt:     types.NewNullableTimeByTime(due),
				EndAt:        types.NewNullableTimeByTime(due.Add(time.Hour)),
			},
			{ // weekly，无 end_at → 续期到次日
				UserId:         100,
				RemindRepeat:   uint8(entities.RemindRepeatWeekly),
				RemindTime:     "09:00",
				RemindWeekdays: weeklyMask,
				RemindAt:       types.NewNullableTimeByTime(due),
			},
			{ // monthly，无 end_at → 续期到次月同日
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatMonthly),
				RemindTime:   "09:00",
				RemindAt:     types.NewNullableTimeByTime(due),
			},
		},
		updated: map[int64]string{},
		cleared: map[int64]bool{},
	}
	// go1.25 语言级别下不能在字面量中引用提升字段 EntityBase.Id，构造后赋值
	for i := range fake.due {
		fake.due[i].Id = int64(i + 1)
	}

	domain := NewTaskDomain(fake, nil)
	tasks, err := domain.ProcessReminders(context.Background())
	if err != nil {
		t.Fatalf("ProcessReminders: %v", err)
	}
	if len(tasks) != 5 {
		t.Fatalf("返回待推送任务数 = %d, want 5", len(tasks))
	}

	wantDaily := due.AddDate(0, 0, 1).Format(time.RFC3339)
	wantWeekly := due.AddDate(0, 0, 1).Format(time.RFC3339)
	wantMonthly := due.AddDate(0, 1, 0).Format(time.RFC3339)

	// 1 daily 无 end_at：续期到次日（H4 核心）
	if got := fake.updated[1]; got != wantDaily {
		t.Fatalf("daily(无 end_at) 应续期到 %s, got %q（被误清空？）", wantDaily, got)
	}
	if fake.cleared[1] {
		t.Error("daily(无 end_at) 不应被清空")
	}
	// 2 一次性：触发后自愈清空，不改期
	if !fake.cleared[2] {
		t.Error("一次性提醒应触发后自愈清空")
	}
	if _, ok := fake.updated[2]; ok {
		t.Error("一次性提醒不应改期")
	}
	// 3 daily 有 end_at 且越过：清空（回归）
	if !fake.cleared[3] {
		t.Error("下一次越过 end_at 的重复提醒应清空")
	}
	if _, ok := fake.updated[3]; ok {
		t.Error("越过 end_at 时不应改期")
	}
	// 4 weekly 无 end_at：续期
	if got := fake.updated[4]; got != wantWeekly {
		t.Fatalf("weekly(无 end_at) 应续期到 %s, got %q", wantWeekly, got)
	}
	// 5 monthly 无 end_at：续期
	if got := fake.updated[5]; got != wantMonthly {
		t.Fatalf("monthly(无 end_at) 应续期到 %s, got %q", wantMonthly, got)
	}
}
