package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// UpdateProject 更新项目值对象
type UpdateProject struct {
	Name        *string
	Description *string
	SortId      *uint16
}

// Validate 验证更新项目值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (updateVO *UpdateProject) Validate() error {
	if updateVO.Name != nil {
		if textutils.RuneLength(*updateVO.Name) == 0 {
			return errors.New("项目名称不能为空")
		}
		if textutils.RuneLength(*updateVO.Name) > 128 {
			return errors.New("项目名称不能超过128个字符")
		}
	}
	if updateVO.Description != nil {
		if textutils.RuneLength(*updateVO.Description) > 512 {
			return errors.New("项目描述不能超过512个字符")
		}
	}
	if updateVO.SortId != nil && *updateVO.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// NewUpdateProject 创建更新项目值对象
// @param name 项目名称
// @param description 项目描述
// @return *UpdateProject 更新项目值对象
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
