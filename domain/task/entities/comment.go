package entities

import (
	"time"
)

type Comment struct {
	Id          int64
	UserId      int64
	TaskId      int64
	Content     string
	Attachments []string
	IsTopUp     bool
	Nickname    string
	Avatar      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
