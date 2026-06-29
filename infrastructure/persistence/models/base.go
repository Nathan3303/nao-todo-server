// Package models 提供持久层数据模型

package models

import (
	"time"

	"gorm.io/gorm"
)

// ModelBase 结构体提供数据模型的基础属性
type ModelBase struct {
	// ID
	// 基于雪花 ID
	ID int64 `gorm:"primaryKey"`

	// 删除时间
	// 通常作为一条记录的兜底删除，而不是直接删除记录
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 创建时间
	// 记录创建时间，Gorm 会自动操作
	CreatedAt time.Time

	// 更新时间
	// 记录更新时间，Gorm 会自动操作
	UpdatedAt time.Time
}

// BeforeCreate Hook
// 雪花 ID 的创建前钩子函数
func (mb *ModelBase) BeforeCreate(tx *gorm.DB) error {
	mb.ID = SnowNode.Generate().Int64()
	return nil
}
