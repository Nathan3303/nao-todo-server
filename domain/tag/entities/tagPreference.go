package entities

import (
	"naotodoserver/domain/types"
)

// TagPreference 标签偏好实体
// 用于表示用户在系统中的标签偏好记录
// 包含用户 ID、标签 ID、视图类型、获取选项、列等属性
type TagPreference struct {
	types.EntityBase
	UserId     types.UserID
	TagId      types.TagID
	ViewType   string
	GetOptions string
	Columns    string
}
