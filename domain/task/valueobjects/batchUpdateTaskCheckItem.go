package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// BatchUpdateTaskCheckItem 批量更新检查项值对象
type BatchUpdateTaskCheckItem struct {
	Id          int64
	Name        *string
	Description *string
	IsDone      *bool
	SortId      *uint16
}

// Validate 验证批量更新事件项值对象是否有效
// @param batchUpdateEvent 批量更新事件项值对象
// @return error 错误信息
func (batchUpdateEvent *BatchUpdateTaskCheckItem) Validate() error {
	if batchUpdateEvent.Id <= 0 {
		return errors.New("任务检查项 ID 无效")
	}
	if batchUpdateEvent.Name != nil && textutils.RuneLength(*batchUpdateEvent.Name) > 128 {
		return errors.New("任务检查项名称长度不能超过 128 个字符")
	}
	if batchUpdateEvent.Description != nil &&
		textutils.RuneLength(*batchUpdateEvent.Description) > 512 {
		return errors.New("任务检查项描述长度不能超过 512 个字符")
	}
	if batchUpdateEvent.SortId != nil && *batchUpdateEvent.SortId == 0 {
		return errors.New("排序 ID 不能为空")
	}
	return nil
}

// NewBatchUpdateTaskCheckItem 创建批量更新任务检查项值对象
// @param id 检查项 ID
// @param name 检查项名称
// @param description 检查项描述
// @param isDone 是否完成
// @param sortId 排序 ID
// @return *BatchUpdateTaskCheckItem 批量更新检查项值对象
// @return error 错误信息
func NewBatchUpdateTaskCheckItem(
	id int64,
	name *string,
	description *string,
	isDone *bool,
	sortId *uint16,
) (*BatchUpdateTaskCheckItem, error) {
	vo := &BatchUpdateTaskCheckItem{
		Id:          id,
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

// BatchUpdateTaskCheckItems 批量更新任务检查项集合
type BatchUpdateTaskCheckItems []*BatchUpdateTaskCheckItem

// Validate 验证批量更新任务检查项集合是否有效
// @return error 错误信息
func (batchUpdateEvents BatchUpdateTaskCheckItems) Validate() error {
	if len(batchUpdateEvents) == 0 {
		return errors.New("任务检查项列表不能为空")
	}
	for _, event := range batchUpdateEvents {
		if err := event.Validate(); err != nil {
			return err
		}
	}
	return nil
}
