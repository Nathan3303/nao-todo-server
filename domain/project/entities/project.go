package entities

import (
	"naotodoserver/domain/types"
)

// Project 任务清单实体
// 用于表示用户在系统中的任务清单记录
// 包含用户 ID、名称、描述、归档时间、创建时间、更新时间、删除时间、停用时间、排序 ID等属性
type Project struct {
	types.EntityBase
	UserId      types.UserID
	Name        string
	Description string
	ArchivedAt  types.NullableTime
	DeactivedAt types.NullableTime
	SortId      uint16
}
