package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// CreateCheckItem 创建事件值对象
type CreateCheckItem struct {
	UserId      int64
	TaskId      int64
	Name        string
	Description string
	SortId      uint16
}

// Validate 验证创建事件值对象是否有效
// @return error 错误信息
func (createEvent *CreateCheckItem) Validate() error {
	if createEvent.TaskId <= 0 {
		return errors.New("任务 ID 不能为空")
	}
	if createEvent.Name == "" {
		return errors.New("事件名称不能为空")
	}
	if textutils.RuneLength(createEvent.Name) > 128 {
		return errors.New("事件名称长度不能超过 128 个字符")
	}
	if createEvent.Description != "" && textutils.RuneLength(createEvent.Description) > 512 {
		return errors.New("事件描述长度不能超过 512 个字符")
	}
	return nil
}

// NewCreateCheckItem 创建创建事件值对象
// @param userId 用户ID
// @param taskId 任务 ID
// @param name 事件名称
// @param description 事件描述
// @return *CreateCheckItem 创建事件值对象
// @return error 错误信息
func NewCreateCheckItem(
	userId int64,
	taskId int64,
	name string,
	description string,
) (*CreateCheckItem, error) {
	vo := &CreateCheckItem{
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
