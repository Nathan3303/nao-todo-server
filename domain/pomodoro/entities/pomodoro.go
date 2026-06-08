package entities

import "time"

type Pomodoro struct {
	Id          int64
	UserId      int64
	SessionId   string
	Type        uint8
	TaskId      int64
	TaskName    string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Duration    int
	Note        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}
