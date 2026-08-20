package entities

import (
	"testing"
	"time"

	"naotodoserver/domain/types"
)

func TestTaskChangeState(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		initialState  TaskState
		nextState     TaskState
		wantState     TaskState
		wantErr       bool
		wantCompleted bool // 进入 Completed 后 CompletedAt 应有效
		keepCompleted bool // 同状态迁移时不应重设完成时间
		wantCleared   bool // 离开 Completed 后 CompletedAt 应清空
	}{
		{name: "pending 到 in-progress", initialState: TaskStatePending, nextState: TaskStateInProgress, wantState: TaskStateInProgress},
		{name: "in-progress 到 completed 写入完成时间", initialState: TaskStateInProgress, nextState: TaskStateCompleted, wantState: TaskStateCompleted, wantCompleted: true},
		{name: "pending 到 completed 写入完成时间", initialState: TaskStatePending, nextState: TaskStateCompleted, wantState: TaskStateCompleted, wantCompleted: true},
		{name: "completed 到 in-progress 清空完成时间", initialState: TaskStateCompleted, nextState: TaskStateInProgress, wantState: TaskStateInProgress, wantCleared: true},
		{name: "completed 到 pending 清空完成时间", initialState: TaskStateCompleted, nextState: TaskStatePending, wantState: TaskStatePending, wantCleared: true},
		{name: "pending 到 pending 幂等", initialState: TaskStatePending, nextState: TaskStatePending, wantState: TaskStatePending},
		{name: "completed 到 completed 幂等保留完成时间", initialState: TaskStateCompleted, nextState: TaskStateCompleted, wantState: TaskStateCompleted, wantCompleted: true, keepCompleted: true},
		{name: "非法状态 0", initialState: TaskStatePending, nextState: TaskState(0), wantState: TaskStatePending, wantErr: true},
		{name: "非法状态 99", initialState: TaskStatePending, nextState: TaskState(99), wantState: TaskStatePending, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{State: tt.initialState}
			if tt.initialState == TaskStateCompleted {
				task.CompletedAt = types.NewNullableTimeByTime(now)
			}
			err := task.ChangeState(tt.nextState)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeState(%d) err = %v, wantErr %v", uint8(tt.nextState), err, tt.wantErr)
			}
			if task.State != tt.wantState {
				t.Errorf("ChangeState 后 State = %v, want %v", task.State, tt.wantState)
			}
			if tt.wantCompleted {
				v, ok := task.CompletedAt.Value()
				if !ok {
					t.Errorf("进入 Completed 后 CompletedAt 应有效")
				}
				if tt.keepCompleted && !v.Equal(now) {
					t.Errorf("同状态迁移不应重设完成时间")
				}
			}
			if tt.wantCleared {
				if _, ok := task.CompletedAt.Value(); ok {
					t.Errorf("离开 Completed 后 CompletedAt 应清空")
				}
			}
		})
	}
}

func TestTaskArchive(t *testing.T) {
	task := &Task{}
	task.Archive(nil)
	if _, ok := task.ArchivedAt.Value(); !ok {
		t.Errorf("Archive 后 ArchivedAt 应有效")
	}
	// 幂等重设时间
	task.Archive(nil)
	if _, ok := task.ArchivedAt.Value(); !ok {
		t.Errorf("重复 Archive 后 ArchivedAt 仍应有效")
	}
	task.Unarchive()
	if _, ok := task.ArchivedAt.Value(); ok {
		t.Errorf("Unarchive 后 ArchivedAt 应清空")
	}
}

func TestTaskUnarchive(t *testing.T) {
	task := &Task{}
	// 先归档再取消归档
	task.Archive(nil)
	if _, ok := task.ArchivedAt.Value(); !ok {
		t.Errorf("Archive 后 ArchivedAt 应有效")
	}
	task.Unarchive()
	if _, ok := task.ArchivedAt.Value(); ok {
		t.Errorf("Unarchive 后 ArchivedAt 应清空")
	}
	// 对未归档任务调用 Unarchive 幂等
	task.Unarchive()
	if _, ok := task.ArchivedAt.Value(); ok {
		t.Errorf("重复 Unarchive 后 ArchivedAt 仍应为空")
	}
}

func TestTaskToggleStar(t *testing.T) {
	task := &Task{}
	// 显式指定时间设置收藏
	task.ToggleStar(nil)
	if _, ok := task.StarMarkAt.Value(); !ok {
		t.Errorf("ToggleStar(nil) 后 StarMarkAt 应有效")
	}
	// 再次 ToggleStar 视为重设收藏时间
	task.ToggleStar(nil)
	if _, ok := task.StarMarkAt.Value(); !ok {
		t.Errorf("再次 ToggleStar(nil) 后 StarMarkAt 仍应有效")
	}
	// 外部清空收藏
	task.StarMarkAt = types.NewNullableTimeNull()
	if _, ok := task.StarMarkAt.Value(); ok {
		t.Errorf("外部清空后 StarMarkAt 应为空")
	}
}

func TestTaskGiveUp(t *testing.T) {
	task := &Task{}
	task.GiveUp(nil)
	if _, ok := task.GivenUpAt.Value(); !ok {
		t.Errorf("GiveUp 后 GivenUpAt 应有效")
	}
	// 幂等重设时间
	task.GiveUp(nil)
	if _, ok := task.GivenUpAt.Value(); !ok {
		t.Errorf("重复 GiveUp 后 GivenUpAt 仍应有效")
	}
	task.UngiveUp()
	if _, ok := task.GivenUpAt.Value(); ok {
		t.Errorf("UngiveUp 后 GivenUpAt 应清空")
	}
}

func TestTaskUngiveUp(t *testing.T) {
	task := &Task{}
	// 先放弃再取消放弃
	task.GiveUp(nil)
	if _, ok := task.GivenUpAt.Value(); !ok {
		t.Errorf("GiveUp 后 GivenUpAt 应有效")
	}
	task.UngiveUp()
	if _, ok := task.GivenUpAt.Value(); ok {
		t.Errorf("UngiveUp 后 GivenUpAt 应清空")
	}
	// 对未放弃任务调用 UngiveUp 幂等
	task.UngiveUp()
	if _, ok := task.GivenUpAt.Value(); ok {
		t.Errorf("重复 UngiveUp 后 GivenUpAt 仍应为空")
	}
}

// TestTaskUnstar 取消收藏：清空收藏时间，且不再校验与开始时间的关系
func TestTaskUnstar(t *testing.T) {
	now := time.Now()
	startAt := now.Add(24 * time.Hour) // 开始时间在未来
	task := &Task{
		StarMarkAt: types.NewNullableTimeByTime(now),
		StartAt:    types.NewNullableTimeByTime(startAt),
	}
	// 收藏时间早于开始时间时，收藏状态本身无效
	if task.IsStarMarkAtValid() {
		t.Fatal("收藏时间早于开始时间时应判无效")
	}
	task.Unstar()
	if _, ok := task.StarMarkAt.Value(); ok {
		t.Fatal("Unstar 后 StarMarkAt 应清空")
	}
	if !task.IsStarMarkAtValid() {
		t.Fatal("Unstar 后不应再校验与开始时间的关系，应判有效")
	}
}
