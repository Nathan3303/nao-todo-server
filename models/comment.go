package models

type Comment struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`
	TodoId int64 `gorm:"not null" json:"todoId"`

	Content     string   `gorm:"size:512" json:"content"`
	Attachments []string `gorm:"serializer:json;type:json" json:"attachments"`
	IsTopUp     bool     `json:"isTopUp"`
}
