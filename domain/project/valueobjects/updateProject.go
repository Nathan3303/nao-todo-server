package valueobjects

import "errors"

// 更新任务清单值对象
type UpdateProject struct {
	Name        string
	Description string
}

// 验证更新任务清单值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (updateVO *UpdateProject) Validate() error {
	if updateVO.Name != "" && len(updateVO.Name) > 128 {
		return errors.New("任务清单名称不能超过128个字符")
	}
	if updateVO.Description != "" && len(updateVO.Description) > 256 {
		return errors.New("任务清单描述不能超过256个字符")
	}
	return nil
}

// 创建更新任务清单值对象
// @param name 任务清单名称
// @param description 任务清单描述
// @return *UpdateProject 更新任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func NewUpdateProject(name string, description string) (*UpdateProject, error) {
	vo := &UpdateProject{
		Name:        name,
		Description: description,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
