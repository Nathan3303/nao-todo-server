package entities

import (
	"naotodoserver/domain/types"
)

// TaskComment 评论实体
// 用于表示用户在系统中的评论记录
// 包含用户 ID、任务 ID、内容、附件、是否是充值、昵称、头像等属性
type TaskComment struct {
	types.EntityBase
	UserId      int64
	TaskId      int64
	Content     string
	Attachments []string
	IsTopUp     bool
	Nickname    string
	Avatar      string
}
