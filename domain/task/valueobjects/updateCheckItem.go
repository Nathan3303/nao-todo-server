package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// UpdateCheckItem 更新事件值对象
type UpdateCheckItem struct {
	Name        *string
	Description *string
	IsDone      *bool
	SortId      *uint16
}

// Validate 验证更新事件值对象是否有效
// @return error 错误信息
func (updateEvent *UpdateCheckItem) Validate() error {
	if updateEvent.Name != nil && textutils.RuneLength(*updateEvent.Name) > 128 {
		return errors.New("事件名称长度不能超过 128 个字符")
	}
	if updateEvent.Description != nil && textutils.RuneLength(*updateEvent.Description) > 512 {
		return errors.New("事件描述长度不能超过 512 个字符")
	}
	if updateEvent.SortId != nil && *updateEvent.SortId == 0 {
		return errors.New("排序 ID 不能为空")
	}
	return nil
}

// NewUpdateCheckItem 创建更新事件值对象
// @param name 事件名称
// @param description 事件描述
// @param isDone 是否完成
// @param sortId 排序 ID
// @return *UpdateCheckItem 更新事件值对象
// @return error 错误信息
func NewUpdateCheckItem(
	name *string,
	description *string,
	isDone *bool,
	sortId *uint16,
) (*UpdateCheckItem, error) {
	vo := &UpdateCheckItem{
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
