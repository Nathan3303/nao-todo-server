package valueobjects

import "errors"

// 创建任务清单值对象
type CreateProject struct {
	UserId      int64
	Name        string
	Description string
	SortId      uint16
}

// 验证创建任务清单值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (createVO *CreateProject) Validate() error {
	// 验证用户 ID 是否为空
	if createVO.UserId == 0 {
		return errors.New("用户 ID 不能为空")
	}
	// 验证任务清单名称是否为空
	if createVO.Name == "" {
		return errors.New("任务清单名称不能为空")
	}
	// 验证任务清单名称长度是否超过128个字符
	if len(createVO.Name) > 128 {
		return errors.New("任务清单名称不能超过128个字符")
	}
	// 验证任务清单描述是否超过256个字符
	if createVO.Description != "" && len(createVO.Description) > 256 {
		return errors.New("任务清单描述不能超过256个字符")
	}
	return nil
}

// 创建任务清单值对象
// @param userId 用户 ID
// @param name 任务清单名称
// @param description 任务清单描述
// @return *CreateProject 创建任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func NewCreateProject(userId int64, name string, description string) (*CreateProject, error) {
	vo := &CreateProject{
		UserId:      userId,
		Name:        name,
		Description: description,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
