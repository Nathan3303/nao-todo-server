package entities

import (
	"naotodoserver/domain/comment/vo"
	"time"
)

type Comment struct {
	Id          int64
	UserId      int64
	TaskId      int64
	Content     string
	Attachments []string
	CommentUser *vo.CommentUser
	CreatedAt   *time.Time
	IsTopUp     bool
}
