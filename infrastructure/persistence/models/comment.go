package models

type Comment struct {
	ModelBase
	UserId      int64    `gorm:"not null;index:idx_comment_user_id"`
	TaskId      int64    `gorm:"not null;index:idx_comment_task_id"`
	Content     string   `gorm:"size:512"`
	Attachments []string `gorm:"serializer:json;type:json"`
	IsTopUp     bool     `gorm:"index:idx_comment_is_top"`
	Nickname    string
	Avatar      string
}
