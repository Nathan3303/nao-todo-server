package models

type Event struct {
	ModelBase
	UserId      int64  `gorm:"not null"`
	TaskId      int64  `gorm:"not null"`
	Name        string `gorm:"size:128"`
	Description string `gorm:"size:512"`
	IsDone      bool
	SortId      int32
}
