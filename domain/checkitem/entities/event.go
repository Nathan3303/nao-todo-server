package entities

import (
	"errors"
	"time"

	"naotodoserver/domain/textutils"
)

type CheckItem struct {
	Id          int64
	UserId      int64
	TaskId      int64
	Name        string
	Description string
	IsDone      bool
	SortId      uint16
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MarkDone 标记为已完成
func (ci *CheckItem) MarkDone() {
	ci.IsDone = true
}

// MarkUndone 标记为未完成
func (ci *CheckItem) MarkUndone() {
	ci.IsDone = false
}

// IsCompleted 是否已完成
func (ci *CheckItem) IsCompleted() bool {
	return ci.IsDone
}

// ToggleDone 切换完成状态
func (ci *CheckItem) ToggleDone() {
	ci.IsDone = !ci.IsDone
}

// Validate 校验实体自身不变性
func (ci *CheckItem) Validate() error {
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
