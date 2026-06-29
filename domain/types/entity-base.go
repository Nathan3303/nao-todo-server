package types

import "time"

// EntityBase 实体基础属性
// 包含实体的 ID、创建时间、更新时间以及删除时间
type EntityBase struct {
	Id        int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt NullableTime
}
