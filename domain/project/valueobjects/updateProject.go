package valueobjects

import (
	"errors"

	"naotodoserver/infrastructure/utils"
)

// 更新任务清单值对象
type UpdateProject struct {
	Name        *string
	Description *string
	SortId      *uint16
}

// 验证更新任务清单值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (updateVO *UpdateProject) Validate() error {
	if updateVO.Name != nil {
		if utils.RuneLength(*updateVO.Name) == 0 {
			return errors.New("任务清单名称不能为空")
		}
		if utils.RuneLength(*updateVO.Name) > 128 {
			return errors.New("任务清单名称不能超过128个字符")
		}
	}
	if updateVO.Description != nil {
		if utils.RuneLength(*updateVO.Description) > 512 {
			return errors.New("任务清单描述不能超过512个字符")
		}
	}
	if updateVO.SortId != nil && *updateVO.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// 创建更新任务清单值对象
// @param name 任务清单名称
// @param description 任务清单描述
// @return *UpdateProject 更新任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func NewUpdateProject(name *string, description *string, sortId *uint16) (*UpdateProject, error) {
	vo := &UpdateProject{
		Name:        name,
		Description: description,
		SortId:      sortId,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
