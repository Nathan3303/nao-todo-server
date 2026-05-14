package valueobjects

import "errors"

// 批量更新标签值对象
type BatchUpdateTag struct {
	Id          int64
	Name        *string
	Description *string
	Color       *string
	SortId      *uint16
}

// 验证批量更新标签值对象
func (vo *BatchUpdateTag) Validate() error {
	if vo.Id <= 0 {
		return errors.New("标签 ID 无效")
	}
	if vo.Name != nil && len(*vo.Name) > 128 {
		return errors.New("标签名称长度不能超过128个字符")
	}
	if vo.Description != nil && len(*vo.Description) > 256 {
		return errors.New("标签描述长度不能超过256个字符")
	}
	if vo.Color != nil && len(*vo.Color) > 16 {
		return errors.New("标签颜色长度不能超过16个字符")
	}
	if vo.SortId != nil && *vo.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// 创建批量更新标签值对象
func NewBatchUpdateTag(id int64, name, description, color *string, sortId *uint16) (*BatchUpdateTag, error) {
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
