package entities

import "time"

type Event struct {
	Id          int64
	UserId      int64
	TaskId      int64
	Name        string
	Description string
	IsDone      bool
	SortId      uint16
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
