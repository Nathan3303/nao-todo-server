package valueobjects

import "errors"

// UpdateEvent 更新事件值对象
type UpdateEvent struct {
	Name        string
	Description string
	IsDone      bool
	SortId      uint16
}

// Validate 验证更新事件值对象是否有效
// @return error 错误信息
func (updateEvent *UpdateEvent) Validate() error {
	if updateEvent.Name != "" && len(updateEvent.Name) > 128 {
		return errors.New("事件名称长度不能超过 128 个字符")
	}
	if updateEvent.Description != "" && len(updateEvent.Description) > 256 {
		return errors.New("事件描述长度不能超过 512 个字符")
	}
	if updateEvent.SortId == 0 {
		return errors.New("排序 ID 不能为空")
	}
	return nil
}

// NewUpdateEvent 创建更新事件值对象
// @param name 事件名称
// @param description 事件描述
// @param isDone 是否完成
// @param sortId 排序 ID
// @return *UpdateEvent 更新事件值对象
// @return error 错误信息
func NewUpdateEvent(
	name string,
	description string,
	isDone bool,
	sortId uint16,
) (*UpdateEvent, error) {
	vo := &UpdateEvent{
		Name:        name,
		Description: description,
		IsDone:      isDone,
		SortId:      sortId,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
