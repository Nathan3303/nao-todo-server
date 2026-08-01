package entities

import (
	"errors"
	"time"

	"naotodoserver/domain/types"
)

// Task 任务实体
// 用于表示用户在系统中的任务记录
// 包含用户 ID、父任务 ID、名称、描述、状态、优先级、开始时间、结束时间、归档时间、收藏时间、放弃时间、
// 项目 ID、标签、提醒时间、提醒重复、提醒时间、提醒周几等属性
type Task struct {
	types.EntityBase
	UserId         types.UserID
	ParentTaskId   types.TaskID
	Name           string
	Description    string
	State          TaskState
	Priority       TaskPriority
	StartAt        types.NullableTime
	EndAt          types.NullableTime
	ProjectId      types.ProjectID
	Tags           []string
	ArchivedAt     types.NullableTime
	StarMarkAt     types.NullableTime
	GivenUpAt      types.NullableTime
	CompletedAt    types.NullableTime
	RemindAt       types.NullableTime
	RemindRepeat   uint8
	RemindTime     string
	RemindWeekdays uint8
	SortId         uint16
}

// IsEndAtValid 检查结束时间是否有效
func (task *Task) IsEndAtValid() bool {
	if task.EndAt.IsNull {
		return false
	}
	if task.StartAt.IsNull {
		return true
	}
	return task.EndAt.Time.After(task.StartAt.Time)
}

// IsArchivedAtValid 检查归档时间是否有效
func (task *Task) IsArchivedAtValid() bool {
	if task.ArchivedAt.IsNull {
		return true
	}
	if task.StartAt.IsNull {
		return true
	}
	return task.ArchivedAt.Time.After(task.StartAt.Time)
}

// IsStarMarkAtValid 检查收藏时间是否有效
func (task *Task) IsStarMarkAtValid() bool {
	if task.StarMarkAt.IsNull {
		return true
	}
	if task.StartAt.IsNull {
		return true
	}
	return task.StarMarkAt.Time.After(task.StartAt.Time)
}

// IsGivenUpAtValid 检查放弃时间是否有效
func (task *Task) IsGivenUpAtValid() bool {
	if task.GivenUpAt.IsNull {
		return true
	}
	if task.StartAt.IsNull {
		return true
	}
	return task.GivenUpAt.Time.After(task.StartAt.Time)
}

// IsDatesValid 检查时间参数是否有效
func (task *Task) IsDatesValid() error {
	if !task.IsEndAtValid() {
		return errors.New("时间参数无效 - 结束时间必须晚于开始时间")
	}
	if !task.IsArchivedAtValid() {
		return errors.New("时间参数无效 - 归档时间必须晚于开始时间")
	}
	if !task.IsStarMarkAtValid() {
		return errors.New("时间参数无效 - 收藏时间必须晚于开始时间")
	}
	if !task.IsGivenUpAtValid() {
		return errors.New("时间参数无效 - 放弃时间必须晚于开始时间")
	}
	return nil
}

// ChangeState 变更任务状态（状态机）
// 迁移规则：
// - 进入已完成（Completed）时写入完成时间
// - 离开已完成时清空完成时间
// - 同状态迁移幂等返回 nil
// @param next 目标状态
// @return error 非法状态返回错误
func (task *Task) ChangeState(next TaskState) error {
	if next.String() == "" {
		return errors.New("非法的任务状态")
	}
	if task.State == next {
		return nil
	}
	if next == TaskStateCompleted {
		task.CompletedAt = types.NewNullableTimeByTime(time.Now())
	}
	if task.State == TaskStateCompleted {
		task.CompletedAt = types.NewNullableTimeNull()
	}
	task.State = next
	return nil
}

// Archive 归档任务（幂等重设归档时间）
func (task *Task) Archive() {
	task.ArchivedAt = types.NewNullableTimeByTime(time.Now())
}

// Unarchive 取消归档任务
func (task *Task) Unarchive() {
	task.ArchivedAt = types.NewNullableTimeNull()
}

// ToggleStar 切换收藏状态
// 已收藏时取消收藏，未收藏时收藏
func (task *Task) ToggleStar() {
	if _, ok := task.StarMarkAt.Value(); ok {
		task.StarMarkAt = types.NewNullableTimeNull()
	} else {
		task.StarMarkAt = types.NewNullableTimeByTime(time.Now())
	}
}
