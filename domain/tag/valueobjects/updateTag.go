package valueobjects

import "errors"

// 更新标签值对象
type UpdateTag struct {
	Name        *string
	Description *string
	Color       *string
}

// 校验更新标签值对象
// @return error 校验失败返回错误，否则返回 nil
func (updateTag *UpdateTag) Validate() error {
	if updateTag.Name != nil && len(*updateTag.Name) > 128 {
		return errors.New("标签名称长度不能超过128个字符")
	}
	if updateTag.Description != nil && len(*updateTag.Description) > 256 {
		return errors.New("标签描述长度不能超过256个字符")
	}
	if updateTag.Color != nil && len(*updateTag.Color) > 16 {
		return errors.New("标签颜色长度不能超过16个字符")
	}
	return nil
}

// 更新标签值对象
// @param name 标签名称
// @param description 标签描述
// @param color 标签颜色
// @return *UpdateTag 更新标签值对象
// @return error 校验失败返回错误，否则返回 nil
func NewUpdateTag(name, description, color *string) (*UpdateTag, error) {
	vo := &UpdateTag{
		Name:        name,
		Description: description,
		Color:       color,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
