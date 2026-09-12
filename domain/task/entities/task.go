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

	// --- 领域统计属性（查询结果承载；服务端 owned 反规范化列，ADR 2026-09-12） ---
	CheckItemCount uint // 检查项数量（含已完成，口径 §6c）
	CommentCount   uint // 评论数量（不含已删除，口径 §6d）
	SubtaskCount   uint // 直接子任务数量（1 层，口径 §6b）
}

// IsEndAtValid 校验截止时间与开始时间的先后（仅在两者都设置时有意义）：
// 未设截止时间（或已清空）视为无需校验，与 ArchivedAt/StarMarkAt/GivenUpAt 的缺席语义一致。
func (task *Task) IsEndAtValid() bool {
	if task.EndAt.IsNull {
		return true
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
		// 离开已完成 = 显式清空完成时间（区别于"从未设置"），持久化层据此写 NULL
		task.CompletedAt = types.NewNullableTimeSetToNull()
	}
	task.State = next
	return nil
}

// Archive 归档任务（可指定归档时间，nil 表示使用当前时间）
func (task *Task) Archive(at *time.Time) {
	if at == nil {
		now := time.Now()
		task.ArchivedAt = types.NewNullableTimeByTime(now)
		return
	}
	task.ArchivedAt = types.NewNullableTimeByTime(*at)
}

// Unarchive 取消归档任务
func (task *Task) Unarchive() {
	task.ArchivedAt = types.NewNullableTimeNull()
}

// ToggleStar 设置收藏时间（可指定收藏时间，nil 表示使用当前时间）
// 由调用方决定是收藏还是取消收藏，本方法仅做赋值
func (task *Task) ToggleStar(at *time.Time) {
	if at == nil {
		now := time.Now()
		task.StarMarkAt = types.NewNullableTimeByTime(now)
		return
	}
	task.StarMarkAt = types.NewNullableTimeByTime(*at)
}

// Unstar 取消收藏任务
// 清空收藏时间（置 null），取消收藏后不再参与与开始时间的关系校验
func (task *Task) Unstar() {
	task.StarMarkAt = types.NewNullableTimeNull()
}

// GiveUp 放弃任务（可指定放弃时间，nil 表示使用当前时间）
func (task *Task) GiveUp(at *time.Time) {
	if at == nil {
		now := time.Now()
		task.GivenUpAt = types.NewNullableTimeByTime(now)
		return
	}
	task.GivenUpAt = types.NewNullableTimeByTime(*at)
}

// UngiveUp 取消放弃任务
func (task *Task) UngiveUp() {
	task.GivenUpAt = types.NewNullableTimeNull()
}
