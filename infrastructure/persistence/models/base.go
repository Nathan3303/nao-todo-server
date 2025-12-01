package models

import (
	"time"

	"gorm.io/gorm"
)

type ModelBase struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (mb *ModelBase) BeforeCreate(tx *gorm.DB) error {
	mb.ID = SnowNode.Generate().Int64()
	return nil
}
