package models

import "time"

type Session struct {
	ModelBase
	UserId     int64     `gorm:"not null;index:idx_session_user_id"`
	Token      string    `gorm:"size:512;index:idx_session_token"`
	ExpiredAt  time.Time `gorm:"not null;index:idx_session_expired"`
	DeviceType string    `gorm:"size:32"`
	IP4        string    `gorm:"size:64"`
	Region     string    `gorm:"size:64"`
}
