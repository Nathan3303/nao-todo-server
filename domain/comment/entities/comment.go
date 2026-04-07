package entities

import (
	"time"
)

type Comment struct {
	Id          int64
	UserId      int64
	TaskId      int64
	Content     string
	Attachments []string
	IsTopUp     bool
	CommentUser *CommentUser
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CommentUser struct {
	Id        int64
	CommentId int64
	Nickname  string
	Avatar    string
}
