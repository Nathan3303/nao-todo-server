package models

import "time"

type Project struct {
	ModelBase
	UserId      int64              `gorm:"not null"`
	Name        string             `gorm:"size:128"`
	Description string             `gorm:"size:256"`
	ArchivedAt  *time.Time         `gorm:"null"`
	Preference  *ProjectPreference `gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE;"`
}

type ProjectPreference struct {
	ModelBase
	UserId     int64  `gorm:"not null"`
	ProjectId  int64  `gorm:"not null;index"`
	ViewType   string `gorm:"size:16"`
	GetOptions string `gorm:"size:256"`
	Columns    string `gorm:"size:256"`
}
