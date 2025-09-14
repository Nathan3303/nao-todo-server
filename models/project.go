package models

import (
	"time"
)

type Project struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`

	Name        string     `gorm:"size:128" json:"name"`
	Description string     `gorm:"size:256" json:"description"`
	ArchivedAt  *time.Time `gorm:"null" json:"archivedAt"`

	Preference *ProjectPreference `gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE;" json:"preference"`
}

type ProjectPreference struct {
	Model

	ProjectId int64 `gorm:"not null;index" json:"projectId"`

	ViewType   string `gorm:"size:16" json:"viewType"`
	GetOptions string `gorm:"size:256" json:"getTodosOptions"`
	Columns    string `gorm:"size:256" json:"columns" default:"priority,project,description,endAt"`
}
