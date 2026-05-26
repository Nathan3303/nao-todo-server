package models

import (
	"database/sql"
)

type Task struct {
	ModelBase
	UserId         int64        `gorm:"not null;index:idx_task_user_id"`
	ParentTaskId   int64        `gorm:"null;index:idx_task_parent"`
	Name           string       `gorm:"size:256"`
	Description    string       `gorm:"size:512"`
	State          int8         `gorm:"default:0;index:idx_task_state"`
	Priority       int8         `gorm:"default:0"`
	StartAt        sql.NullTime `gorm:"null"`
	EndAt          sql.NullTime `gorm:"null;index:idx_task_end_at"`
	ProjectId      int64        `gorm:"not null;index:idx_task_project"`
	Tags           []string     `gorm:"serializer:json;type:json"`
	ArchivedAt     sql.NullTime `gorm:"null"`
	StarMarkAt     sql.NullTime `gorm:"null"`
	GivenUpAt      sql.NullTime `gorm:"null"`
	RemindAt       sql.NullTime `gorm:"null;index:idx_task_remind_at"`
	RemindRepeat   int8         `gorm:"default:0"`
	RemindTime     string       `gorm:"size:5"`
	RemindWeekdays int8         `gorm:"default:0"`
}
