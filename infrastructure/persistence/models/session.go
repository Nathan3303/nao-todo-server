package models

import "time"

type Session struct {
	ModelBase
	UserId     int64     `gorm:"not null"`
	Token      string    `gorm:"size:512;index"`
	ExpiredAt  time.Time `gorm:"not null"`
	DeviceType string    `gorm:"size:32"`
}
