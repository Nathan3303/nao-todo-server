package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// BatchUpdateTag 批量更新标签值对象
type BatchUpdateTag struct {
	Id          int64
	Name        *string
	Description *string
	Color       *string
	SortId      *uint16
}

// Validate 验证批量更新标签值对象是否符合要求
func (vo *BatchUpdateTag) Validate() error {
	if vo.Id <= 0 {
		return errors.New("标签 ID 无效")
	}
	if vo.Name != nil && textutils.RuneLength(*vo.Name) > 64 {
		return errors.New("标签名称长度不能超过64个字符")
	}
	if vo.Description != nil && textutils.RuneLength(*vo.Description) > 512 {
		return errors.New("标签描述长度不能超过512个字符")
	}
	if vo.Color != nil && textutils.RuneLength(*vo.Color) > 16 {
		return errors.New("标签颜色长度不能超过16个字符")
	}
	if vo.SortId != nil && *vo.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// NewBatchUpdateTag 创建批量更新标签值对象
func NewBatchUpdateTag(
	id int64,
	name *string,
	description *string,
	color *string,
	sortId *uint16,
) (*BatchUpdateTag, error) {
	vo := &BatchUpdateTag{
		Id:          id,
		Name:        name,
		Description: description,
		Color:       color,
		SortId:      sortId,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
