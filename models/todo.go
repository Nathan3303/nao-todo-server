package models

import "time"

type Todo struct {
	Model

	UserId       int64  `gorm:"not null" json:"userId"`
	ProjectId    int64  `gorm:"not null" json:"projectId"`
	ParentTodoId *int64 `gorm:"null" json:"parentTodoId"`

	Name        string     `gorm:"size:64" json:"name"`
	Description string     `gorm:"size:256" json:"description"`
	State       int8       `json:"state"`
	Priority    int8       `json:"priority"`
	StartAt     *time.Time `gorm:"null" json:"startAt"`
	EndAt       *time.Time `gorm:"null" json:"endAt"`
	ArchivedAt  *time.Time `gorm:"null" json:"archivedAt"`
	FavoritedAt *time.Time `json:"favoritedAt"`
	GivenUpAt   *time.Time `json:"givenUpAt"`
	Tags        []int64    `gorm:"serializer:json;type:json" json:"tags"`
}
