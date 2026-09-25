package entities

import (
	"naotodoserver/domain/types"
)

// Tag 标签实体
// 用于表示用户在系统中的标签记录
// 包含用户 ID、名称、描述、颜色、排序 ID等属性
type Tag struct {
	types.EntityBase
	UserId      types.UserID
	Name        string
	Description string
	Color       string
	SortId      uint16
}
