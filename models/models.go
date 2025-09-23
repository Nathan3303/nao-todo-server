package models

import (
	"naotodoserver/globals"
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type ModelResponse struct {
	ID        string         `json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	m.ID = globals.Vars.SnowNode.Generate().Int64()
	return nil
}
