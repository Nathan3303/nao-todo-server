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

// TestProcessReminders_RescheduleAndSelfHeal 重复提醒改期 / 一次性提醒自愈不受口径补齐影响。
// 注：改期路径要求 end_at 有界（既有 calculateNextRemindAt 以 endAt 为终止边界），
// 故用例显式给定 end_at；end_at 为空时的行为不在本次改动范围。
func TestProcessReminders_RescheduleAndSelfHeal(t *testing.T) {
	due := time.Date(2026, 9, 21, 9, 0, 0, 0, time.UTC)
	fake := &fakeRemindRepo{
		due: []*entities.Task{
			{
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatDaily),
				RemindTime:   "09:00",
				RemindAt:     types.NewNullableTimeByTime(due),
				EndAt:        types.NewNullableTimeByTime(due.AddDate(0, 0, 30)),
			},
			{
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatNone),
				RemindAt:     types.NewNullableTimeByTime(due),
			},
			{
				UserId:       100,
				RemindRepeat: uint8(entities.RemindRepeatDaily),
				RemindTime:   "09:00",
				RemindAt:     types.NewNullableTimeByTime(due),
				EndAt:        types.NewNullableTimeByTime(due.Add(time.Hour)),
			},
		},
		updated: map[int64]string{},
		cleared: map[int64]bool{},
	}
	// go1.25 语言级别下不能在字面量中引用提升字段 EntityBase.Id，构造后赋值
	fake.due[0].Id = 1
	fake.due[1].Id = 2
	fake.due[2].Id = 3

	domain := NewTaskDomain(fake, nil)
	tasks, err := domain.ProcessReminders(context.Background())
	if err != nil {
		t.Fatalf("ProcessReminders: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("返回待推送任务数 = %d, want 3", len(tasks))
	}

	// 重复提醒：改期到次日同一时刻，不清空
	wantNext := due.AddDate(0, 0, 1).Format(time.RFC3339)
	if got := fake.updated[1]; got != wantNext {
		t.Fatalf("重复提醒应改期到 %s, got %q", wantNext, got)
	}
	if fake.cleared[1] {
		t.Error("重复提醒不应被清空")
	}

	// 一次性提醒：触发后自愈清空，不改期
	if !fake.cleared[2] {
		t.Error("一次性提醒应触发后自愈清空")
	}
	if _, ok := fake.updated[2]; ok {
		t.Error("一次性提醒不应改期")
	}

	// 重复提醒但下一次已越过 end_at：按既有规则清空
	if !fake.cleared[3] {
		t.Error("下一次越过 end_at 的重复提醒应清空")
	}
}
