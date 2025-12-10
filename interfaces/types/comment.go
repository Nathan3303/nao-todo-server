package types

type CommentRes struct {
	Id          string          `json:"id"`
	TaskId      string          `json:"taskId"`
	Content     string          `json:"content"`
	CreatedAt   string          `json:"createdAt"`
	Attachments []string        `json:"attachments"`
	IsTopUp     bool            `json:"isTopUp"`
	CommentUser *CommentUserRes `json:"commentUser"`
}

type CommentUserRes struct {
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
}

type CreateCommentReq struct {
	TaskId  string `json:"taskId" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type UpdateCommentReq struct {
	Content     string   `json:"content"`
	Attachments []string `json:"attachments"`
	IsTopUp     bool     `json:"isTopUp"`
}

type UpdateCommentRes struct {
	CommentId string `json:"commentId"`
}

type DeleteCommentRes struct {
	CommentId string `json:"commentId"`
}

type ListCommentRes []*CommentRes
