package valueobjects

import "errors"

// 批量更新任务清单值对象
type BatchUpdateProject struct {
	Id          int64
	Name        *string
	Description *string
	SortId      *uint16
}

// 验证批量更新任务清单值对象
func (vo *BatchUpdateProject) Validate() error {
	if vo.Id <= 0 {
		return errors.New("任务清单 ID 无效")
	}
	if vo.Name != nil && len(*vo.Name) > 128 {
		return errors.New("任务清单名称不能超过128个字符")
	}
	if vo.Description != nil && len(*vo.Description) > 256 {
		return errors.New("任务清单描述不能超过256个字符")
	}
	if vo.SortId != nil && *vo.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// 创建批量更新任务清单值对象
func NewBatchUpdateProject(id int64, name, description *string, sortId *uint16) (*BatchUpdateProject, error) {
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
