package entities

import "time"

type TagPreference struct {
	Id         int64
	UserId     int64
	TagId      int64
	ViewType   string
	GetOptions string
	Columns    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
