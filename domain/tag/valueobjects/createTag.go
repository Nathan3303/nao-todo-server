package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// 创建标签值对象
type CreateTag struct {
	Name        string
	Description string
	Color       string
	SortId      uint16
}

// 校验创建标签值对象
// @return error 校验失败返回错误，否则返回 nil
func (createTag *CreateTag) Validate() error {
	if createTag.Name == "" {
		return errors.New("标签名称不能为空")
	}
	if textutils.RuneLength(createTag.Name) > 64 {
		return errors.New("标签名称长度不能超过64个字符")
	}
	if createTag.Description != "" && textutils.RuneLength(createTag.Description) > 512 {
		return errors.New("标签描述长度不能超过512个字符")
	}
	if createTag.Color == "" {
		return errors.New("标签颜色不能为空")
	}
	if textutils.RuneLength(createTag.Color) > 16 {
		return errors.New("标签颜色长度不能超过16个字符")
	}
	return nil
}

// 创建标签值对象
// @param name 标签名称
// @param description 标签描述
// @param color 标签颜色
// @return *CreateTag 创建标签值对象
// @return error 校验失败返回错误，否则返回 nil
func NewCreateTag(name, description, color string) (*CreateTag, error) {
	vo := &CreateTag{
		Name:        name,
		Description: description,
		Color:       color,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
