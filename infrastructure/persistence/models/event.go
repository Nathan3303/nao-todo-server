package models

type Event struct {
	ModelBase
	UserId      int64  `gorm:"not null;index:idx_event_user_id"`
	TaskId      int64  `gorm:"not null;index:idx_event_task_id"`
	Name        string `gorm:"size:128"`
	Description string `gorm:"size:512"`
	IsDone      bool   `gorm:"index:idx_event_is_done"`
	SortId      uint16
}
