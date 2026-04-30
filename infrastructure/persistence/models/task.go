package models

import (
	"database/sql"
)

type Task struct {
	ModelBase
	UserId       int64        `gorm:"not null"`
	ParentTaskId int64        `gorm:"null"`
	Name         string       `gorm:"size:256" `
	Description  string       `gorm:"size:512" `
	State        int8         `gorm:"default:0"`
	Priority     int8         `gorm:"default:0"`
	StartAt      sql.NullTime `gorm:"null"`
	EndAt        sql.NullTime `gorm:"null"`
	ProjectId    int64        `gorm:"not null" `
	Tags         []string     `gorm:"serializer:json;type:json"`
	ArchivedAt   sql.NullTime `gorm:"null"`
	StarMarkAt   sql.NullTime `gorm:"null"`
	GivenUpAt    sql.NullTime `gorm:"null"`
}
