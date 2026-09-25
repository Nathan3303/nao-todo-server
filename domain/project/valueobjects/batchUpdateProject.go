package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// BatchUpdateProject 批量更新项目值对象
type BatchUpdateProject struct {
	Id          int64
	Name        *string
	Description *string
	SortId      *uint16
}

// Validate 验证批量更新项目值对象
func (vo *BatchUpdateProject) Validate() error {
	if vo.Id <= 0 {
		return errors.New("项目 ID 无效")
	}
	if vo.Name != nil && textutils.RuneLength(*vo.Name) > 128 {
		return errors.New("项目名称不能超过128个字符")
	}
	if vo.Description != nil && textutils.RuneLength(*vo.Description) > 512 {
		return errors.New("项目描述不能超过512个字符")
	}
	if vo.SortId != nil && *vo.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// NewBatchUpdateProject 创建批量更新项目值对象
func NewBatchUpdateProject(
	id int64,
	name *string,
	description *string,
	sortId *uint16,
) (*BatchUpdateProject, error) {
	vo := &BatchUpdateProject{
		Id:          id,
		Name:        name,
		Description: description,
		SortId:      sortId,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
