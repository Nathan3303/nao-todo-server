package entities

import (
	"time"

	"gorm.io/gorm"
)

type UserConfig struct {
	Id         int64          `json:"id"`
	UserId     int64          `json:"userId"`
	State      string         `json:"state"`
	Appearance string         `json:"appearance"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `json:"deletedAt"`
}
