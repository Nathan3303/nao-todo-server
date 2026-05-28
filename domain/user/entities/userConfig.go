package entities

import (
	"time"
)

type UserConfig struct {
	Id         int64          `json:"id"`
	UserId     int64          `json:"userId"`
	Appearance string         `json:"appearance"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	DeletedAt  time.Time `json:"deletedAt"`
}
