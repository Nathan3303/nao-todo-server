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
	Tags        []string   `gorm:"serializer:json;type:json" json:"tags"`
}

type TodoResponse struct {
	ModelResponse

	UserId       string  `json:"userId"`
	ProjectId    string  `json:"projectId"`
	ParentTodoId *string `json:"parentTodoId"`

	Name        string     `json:"name"`
	Description string     `json:"description"`
	State       string     `json:"state"`
	Priority    string     `json:"priority"`
	StartAt     *time.Time `json:"startAt"`
	EndAt       *time.Time `json:"endAt"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	FavoritedAt *time.Time `json:"favoritedAt"`
	GivenUpAt   *time.Time `json:"givenUpAt"`
	Tags        []string   `json:"tags"`

	// extras fill
	IsArchived  bool `json:"isArchived"`
	IsDeleted   bool `json:"isDeleted"`
	IsFavorited bool `json:"isFavorited"`
	IsGivenUp   bool `json:"isGivenUp"`
}
