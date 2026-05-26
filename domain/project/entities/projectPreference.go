package entities

import "time"

type ProjectPreference struct {
	Id         int64
	UserId     int64
	ProjectId  int64
	ViewType   string
	GetOptions string
	Columns    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}
