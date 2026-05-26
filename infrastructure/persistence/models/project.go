package models

import "time"

type Project struct {
	ModelBase
	UserId      int64      `gorm:"not null;index:idx_project_user_id"`
	Name        string     `gorm:"size:128"`
	Description string     `gorm:"size:512"`
	ArchivedAt  *time.Time `gorm:"null;index:idx_project_archived_at"`
	DeactivedAt *time.Time `gorm:"index:idx_project_deactived_at"`
	SortId      uint16
	Preference  *ProjectPreference `gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE;"`
}

type ProjectPreference struct {
	ModelBase
	UserId     int64  `gorm:"not null;index:idx_proj_pref_user"`
	ProjectId  int64  `gorm:"not null;index:idx_proj_pref_project"`
	ViewType   string `gorm:"size:16"`
	GetOptions string `gorm:"size:512"`
	Columns    string `gorm:"size:512"`
}
