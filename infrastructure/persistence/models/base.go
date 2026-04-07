package models

import (
	"time"

	"gorm.io/gorm"
)

type ModelBase struct {
	ID        int64          `gorm:"primaryKey"`
	UUID      string         `gorm:"primaryKey"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (mb *ModelBase) BeforeCreate(tx *gorm.DB) error {
	mb.ID = SnowNode.Generate().Int64()
	return nil
}
