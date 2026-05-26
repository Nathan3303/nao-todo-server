package entities

import (
	"time"
)

// 任务清单实体
type Project struct {
	Id          int64
	UserId      int64
	Name        string
	Description string
	ArchivedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	DeactivedAt *time.Time
	SortId      uint16
}
