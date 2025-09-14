package models

import "time"

type Session struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`

	JWT        string    `gorm:"size:512" json:"jwt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	DeviceType string    `gorm:"size:16" json:"deviceType"`
}
