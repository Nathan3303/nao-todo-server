package models

import "time"

type Project struct {
	ModelBase
	UserId      int64              `gorm:"not null" json:"userId"`
	Name        string             `gorm:"size:128" json:"name"`
	Description string             `gorm:"size:256" json:"description"`
	ArchivedAt  *time.Time         `gorm:"null" json:"archivedAt"`
	Preference  *ProjectPreference `gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE;" json:"preference"`
}

type ProjectPreference struct {
	ModelBase
	UserId     int64  `gorm:"not null" json:"userId"`
	ProjectId  int64  `gorm:"not null;index" json:"projectId"`
	ViewType   string `gorm:"size:16" json:"viewType"`
	GetOptions string `gorm:"size:256" json:"getTodosOptions" default:""`
	Columns    string `gorm:"size:256" json:"columns" default:""`
}
