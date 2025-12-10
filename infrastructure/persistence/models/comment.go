package models

type Comment struct {
	ModelBase
	UserId      int64        `gorm:"not null"`
	TaskId      int64        `gorm:"not null"`
	Content     string       `gorm:"size:512"`
	Attachments []string     `gorm:"serializer:json;type:json"`
	CommentUser *CommentUser `gorm:"foreignKey:CommentId"`
	IsTopUp     bool
}

type CommentUser struct {
	ModelBase
	CommentId int64 `gorm:"not null"`
	Avatar    string
	Nickname  string
}
