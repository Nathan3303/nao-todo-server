package entities

import (
	"testing"

	"naotodoserver/domain/types"
)

// newCheckItem 构造带基础字段的检查项实体
func newCheckItem(done bool, sortID uint16) *TaskCheckItem {
	return &TaskCheckItem{
		EntityBase: types.EntityBase{Id: 9001},
		UserId:     1001,
		TaskId:     2001,
		Name:       "检查项",
		IsDone:     done,
		SortId:     sortID,
	}
}

func TestTaskCheckItemSetDone(t *testing.T) {
	tests := []struct {
		name      string
		initial   bool
		desired   bool
		wantState bool
	}{
		{name: "未完成→完成", initial: false, desired: true, wantState: true},
		{name: "完成→未完成", initial: true, desired: false, wantState: false},
		{name: "幂等：已完成→已完成", initial: true, desired: true, wantState: true},
		{name: "幂等：未完成→未完成", initial: false, desired: false, wantState: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := newCheckItem(tt.initial, 300)
			item.SetDone(tt.desired)
			if item.IsDone != tt.wantState {
				t.Fatalf("SetDone(%v) 后 IsDone = %v, 期望 %v", tt.desired, item.IsDone, tt.wantState)
			}
		})
	}
}

// TestTaskCheckItemSetDoneOnlyTouchesIsDone SetDone 不得影响其它字段
func TestTaskCheckItemSetDoneOnlyTouchesIsDone(t *testing.T) {
	item := newCheckItem(false, 300)
	item.SetDone(true)
	if item.Name != "检查项" || item.SortId != 300 || item.TaskId != 2001 || item.UserId != 1001 {
		t.Fatalf("SetDone 应只改变 IsDone，实际实体被污染: %+v", item)
	}
}
