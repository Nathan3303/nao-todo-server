package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/textutils"
)

// UpdateTag 更新标签值对象
type UpdateTag struct {
	Name        *string
	Description *string
	Color       *string
	SortId      *uint16
	// UpdatedAt 乐观锁时间戳（零值=未提供）：早于服务端当前版本时不更新（LWW）
	UpdatedAt time.Time
}

// Validate 校验更新标签值对象
// @return error 校验失败返回错误，否则返回 nil
func (updateTag *UpdateTag) Validate() error {
	if updateTag.Name != nil && textutils.RuneLength(*updateTag.Name) > 64 {
		return errors.New("标签名称长度不能超过64个字符")
	}
	if updateTag.Description != nil && textutils.RuneLength(*updateTag.Description) > 512 {
		return errors.New("标签描述长度不能超过512个字符")
	}
	if updateTag.Color != nil && textutils.RuneLength(*updateTag.Color) > 16 {
		return errors.New("标签颜色长度不能超过16个字符")
	}
	if updateTag.SortId != nil && *updateTag.SortId == 0 {
		return errors.New("排序ID不能为0")
	}
	return nil
}

// NewUpdateTag 更新标签值对象
// @param name 标签名称
// @param description 标签描述
// @param color 标签颜色
// @return *UpdateTag 更新标签值对象
// @return error 校验失败返回错误，否则返回 nil
func NewUpdateTag(name, description, color *string, sortId *uint16) (*UpdateTag, error) {
	vo := &UpdateTag{
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
