package valueobjects

import (
	"errors"
)

// BatchUpdateEvent 批量更新事件值对象
type BatchUpdateEvent struct {
	Id          int64
	Name        *string
	Description *string
	IsDone      *bool
	SortId      *uint16
}

// Validate 验证批量更新事件值对象是否有效
// @param batchUpdateEvent 批量更新事件值对象
// @return error 错误信息
func (batchUpdateEvent *BatchUpdateEvent) Validate() error {
	if batchUpdateEvent.Id <= 0 {
		return errors.New("事件 ID 无效")
	}
	if batchUpdateEvent.Name != nil && len(*batchUpdateEvent.Name) > 128 {
		return errors.New("事件名称长度不能超过 128 个字符")
	}
	if batchUpdateEvent.Description != nil && len(*batchUpdateEvent.Description) > 256 {
		return errors.New("事件描述长度不能超过 512 个字符")
	}
	if batchUpdateEvent.SortId != nil && *batchUpdateEvent.SortId == 0 {
		return errors.New("排序 ID 不能为空")
	}
	return nil
}

// NewBatchUpdateEvent 创建批量更新事件值对象
// @param id 事件 ID
// @param name 事件名称
// @param description 事件描述
// @param isDone 是否完成
// @param sortId 排序 ID
// @return *BatchUpdateEvent 批量更新事件值对象
// @return error 错误信息
func NewBatchUpdateEvent(
	id int64,
	name *string,
	description *string,
	isDone *bool,
	sortId *uint16,
) (*BatchUpdateEvent, error) {
	vo := &BatchUpdateEvent{
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

// BatchUpdateEvents 批量更新事件集合
type BatchUpdateEvents []*BatchUpdateEvent

// Validate 验证批量更新事件集合是否有效
// @return error 错误信息
func (batchUpdateEvents BatchUpdateEvents) Validate() error {
	if len(batchUpdateEvents) == 0 {
		return errors.New("事件列表不能为空")
	}

	for _, event := range batchUpdateEvents {
		if err := event.Validate(); err != nil {
			return err
		}
	}

	return nil
}
