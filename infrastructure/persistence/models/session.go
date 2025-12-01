package models

import "time"

type Session struct {
	ModelBase
	UserId     int64     `gorm:"not null" json:"userId"`
	Token      string    `gorm:"size:512;index" json:"token"`
	ExpiredAt  time.Time `gorm:"not null" json:"expiredAt"`
	DeviceType string    `gorm:"size:32" json:"deviceType"`
}
