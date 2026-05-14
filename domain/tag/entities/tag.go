package entities

import (
	"time"
)

type Tag struct {
	Id          int64
	UserId      int64
	Name        string
	Description string
	Color       string
	SortId      uint16
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
