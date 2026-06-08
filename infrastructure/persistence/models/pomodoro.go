package models

import "database/sql"

type Pomodoro struct {
	ModelBase
	UserId      int64        `gorm:"not null;index:idx_pomodoro_user_id"`
	SessionId   string       `gorm:"size:36;not null;index:idx_pomodoro_session_id"`
	Type        uint8        `gorm:"not null;default:0"`
	TaskId      int64        `gorm:"not null;index:idx_pomodoro_task_id"`
	TaskName    string       `gorm:"size:256;not null"`
	Description string       `gorm:"size:512"`
	StartAt     sql.NullTime `gorm:"null;index:idx_pomodoro_start_at"`
	EndAt       sql.NullTime `gorm:"null"`
	Duration    int          `gorm:"not null"`
	Note        string       `gorm:"type:text"`
}
