package modules

import "time"

type Tag struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`

	Name        string     `gorm:"size:128" json:"name"`
	Description string     `gorm:"size:256" json:"description"`
	ArchivedAt  *time.Time `gorm:"null" json:"archivedAt"`
	Color       string     `gorm:"size:16" json:"color"`

	Preference *TagPreference `gorm:"foreignKey:TagId;constraint:OnDelete:CASCADE;" json:"preference"`
}

type TagPreference struct {
	Model

	TagId int64 `gorm:"not null;index" json:"tagId"`

	ViewType   string `gorm:"size:16" json:"viewType"`
	GetOptions string `gorm:"size:256" json:"getTodosOptions"`
	Columns    string `gorm:"size:256" json:"columns" default:"priority,project,description,endAt"`
}
