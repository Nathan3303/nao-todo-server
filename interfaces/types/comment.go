package types

type CommentRes struct {
	Id          string   `json:"id"`
	TaskId      string   `json:"taskId"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments"`
	IsTopUp     bool     `json:"isTopUp"`
	Nickname    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type CreateCommentReq struct {
	TaskId  string `json:"taskId" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type UpdateCommentReq struct {
	Content     *string   `json:"content"`
	Attachments *[]string `json:"attachments"`
	IsTopUp     *bool     `json:"isTopUp"`
}

type ListCommentRes []*CommentRes
