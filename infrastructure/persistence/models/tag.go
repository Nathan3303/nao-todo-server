package models

type Tag struct {
	ModelBase
	UserId      int64          `gorm:"not null"`
	Name        string         `gorm:"size:128"`
	Description string         `gorm:"size:256"`
	Color       string         `gorm:"size:16"`
	Preference  *TagPreference `gorm:"foreignKey:TagId;constraint:OnDelete:CASCADE;"`
}

type TagPreference struct {
	ModelBase
	UserId     int64  `gorm:"not null"`
	TagId      int64  `gorm:"not null;index"`
	ViewType   string `gorm:"size:16"`
	GetOptions string `gorm:"size:256"`
	Columns    string `gorm:"size:256"`
}
