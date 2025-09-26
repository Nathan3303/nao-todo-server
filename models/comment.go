package models

type Comment struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`
	TodoId int64 `gorm:"not null" json:"todoId"`

	Content     string   `gorm:"size:512" json:"content"`
	Attachments []string `gorm:"serializer:json;type:json" json:"attachments"`
	IsTopUp     bool     `json:"isTopUp"`

	CommentUser *CommentUser `gorm:"foreignKey:CommentId" json:"user"`
}

type CommentUser struct {
	Model

	CommentId int64 `gorm:"not null" json:"commentId"`

	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
}

type CommentUserResponse struct {
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
}

type CommentResponse struct {
	ModelResponse

	UserId string `json:"userId"`
	TodoId string `json:"todoId"`

	Content     string   `json:"content"`
	Attachments []string `json:"attachments"`
	IsTopUp     bool     `json:"isTopUp"`

	CommentUser *CommentUserResponse `json:"user"`
}
