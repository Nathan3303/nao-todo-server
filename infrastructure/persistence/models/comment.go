package models

type Comment struct {
	ModelBase
	UserId      int64        `gorm:"not null;index:idx_comment_user_id"`
	TaskId      int64        `gorm:"not null;index:idx_comment_task_id"`
	Content     string       `gorm:"size:512"`
	Attachments []string     `gorm:"serializer:json;type:json"`
	CommentUser *CommentUser `gorm:"foreignKey:CommentId"`
	IsTopUp     bool         `gorm:"index:idx_comment_is_top"`
}

type CommentUser struct {
	ModelBase
	CommentId int64 `gorm:"not null;index:idx_comment_user_comment"`
	Avatar    string
	Nickname  string
}
