package models

type Tag struct {
	ModelBase
	UserId      int64          `gorm:"not null"`
	Name        string         `gorm:"size:64"`
	Description string         `gorm:"size:512"`
	Color       string         `gorm:"size:16"`
	Preference  *TagPreference `gorm:"foreignKey:TagId;constraint:OnDelete:CASCADE;"`
}

type TagPreference struct {
	ModelBase
	UserId     int64  `gorm:"not null"`
	TagId      int64  `gorm:"not null;index"`
	ViewType   string `gorm:"size:16"`
	GetOptions string `gorm:"size:512"`
	Columns    string `gorm:"size:512"`
}
