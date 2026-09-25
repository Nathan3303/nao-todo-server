package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/textutils"
)

// UpdateTaskCheckItem 更新待办任务检查项值对象
type UpdateTaskCheckItem struct {
	Name        *string
	Description *string
	IsDone      *bool
	SortId      *uint16
	// UpdatedAt 乐观锁时间戳（零值=未提供）：早于服务端当前版本时不更新（LWW）
	UpdatedAt time.Time
}

// Validate 验证更新检查项值对象是否有效
// @return error 错误信息
func (updateCheckItem *UpdateTaskCheckItem) Validate() error {
	if updateCheckItem.Name != nil && textutils.RuneLength(*updateCheckItem.Name) > 128 {
		return errors.New("检查项名称长度不能超过 128 个字符")
	}
	if updateCheckItem.Description != nil &&
		textutils.RuneLength(*updateCheckItem.Description) > 512 {
		return errors.New("检查项描述长度不能超过 512 个字符")
	}
	if updateCheckItem.SortId != nil && *updateCheckItem.SortId == 0 {
		return errors.New("检查项排序 ID 不能为空")
	}
	return nil
}

// NewUpdateTaskCheckItem 创建更新待办任务检查项值对象
// @param name 检查项名称
// @param description 检查项描述
// @param isDone 是否完成
// @param sortId 排序 ID
// @return *UpdateTaskCheckItem 更新待办任务检查项值对象
// @return error 错误信息
func NewUpdateTaskCheckItem(
	name *string,
	description *string,
	isDone *bool,
	sortId *uint16,
) (*UpdateTaskCheckItem, error) {
	vo := &UpdateTaskCheckItem{
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
