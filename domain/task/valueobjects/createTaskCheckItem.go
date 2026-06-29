package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// CreateTaskCheckItem 创建任务检查项值对象
type CreateTaskCheckItem struct {
	UserId      int64
	TaskId      int64
	Name        string
	Description string
	SortId      uint16
}

// Validate 验证创建任务检查项值对象是否有效
// @return error 错误信息
func (createEvent *CreateTaskCheckItem) Validate() error {
	if createEvent.TaskId <= 0 {
		return errors.New("任务 ID 不能为空")
	}
	if createEvent.Name == "" {
		return errors.New("检查项名称不能为空")
	}
	if textutils.RuneLength(createEvent.Name) > 128 {
		return errors.New("检查项名称长度不能超过 128 个字符")
	}
	if createEvent.Description != "" && textutils.RuneLength(createEvent.Description) > 512 {
		return errors.New("检查项描述长度不能超过 512 个字符")
	}
	return nil
}

// NewCreateTaskCheckItem 创建创建任务检查项值对象
// @param userId 用户ID
// @param taskId 任务 ID
// @param name 检查项名称
// @param description 检查项描述
// @return *CreateTaskCheckItem 创建任务检查项值对象
// @return error 错误信息
func NewCreateTaskCheckItem(
	userId int64,
	taskId int64,
	name string,
	description string,
) (*CreateTaskCheckItem, error) {
	vo := &CreateTaskCheckItem{
		UserId:      userId,
		TaskId:      taskId,
		Name:        name,
		Description: description,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
