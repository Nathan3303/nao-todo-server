package entities

import (
	"errors"

	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

// TaskCheckItem 检查项实体
// 用于表示用户在系统中的检查项记录
// 包含用户 ID、任务 ID、名称、描述、是否完成、排序 ID等属性
type TaskCheckItem struct {
	types.EntityBase
	UserId      int64
	TaskId      int64
	Name        string
	Description string
	IsDone      bool
	SortId      uint16
}

// MarkDone 标记为已完成
func (ci *TaskCheckItem) MarkDone() {
	ci.IsDone = true
}

// MarkUndone 标记为未完成
func (ci *TaskCheckItem) MarkUndone() {
	ci.IsDone = false
}

// IsCompleted 是否已完成
func (ci *TaskCheckItem) IsCompleted() bool {
	return ci.IsDone
}

// ToggleDone 切换完成状态
func (ci *TaskCheckItem) ToggleDone() {
	ci.IsDone = !ci.IsDone
}

// Validate 校验实体自身不变性
func (ci *TaskCheckItem) Validate() error {
	if ci.Id <= 0 {
		return errors.New("ID 无效")
	}
	if ci.UserId <= 0 {
		return errors.New("用户 ID 无效")
	}
	if ci.TaskId <= 0 {
		return errors.New("任务 ID 无效")
	}
	if ci.Name == "" {
		return errors.New("名称不能为空")
	}
	if textutils.RuneLength(ci.Name) > 128 {
		return errors.New("名称不能超过128个字符")
	}
	if ci.Description != "" && textutils.RuneLength(ci.Description) > 512 {
		return errors.New("描述不能超过512个字符")
	}
	if ci.SortId == 0 {
		return errors.New("排序 ID 不能为0")
	}
	return nil
}
