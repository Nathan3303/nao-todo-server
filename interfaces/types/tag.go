package types

import "time"

type GetTagRes struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	SortId      uint16    `json:"sortId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GetTagPreferenceRes struct {
	Id         string    `json:"id"`
	TagId      string    `json:"tagId"`
	ViewType   string    `json:"viewType"`
	GetOptions string    `json:"getTasksOptions"`
	Columns    string    `json:"columns"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type CreateTagReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"required"`
}

type CreateTagRes struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	SortId      uint16    `json:"sortId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateTagReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	SortId      *uint16 `json:"sortId"`
}

type BatchUpdateTagReq struct {
	Tags []*struct {
		Id          string  `json:"id" binding:"required"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		SortId      *uint16 `json:"sortId"`
	} `json:"tags" binding:"required,min=1"`
}

type BatchUpdateTagRes struct {
	UpdatedCount int64         `json:"updatedCount"`
	Tags         []*GetTagRes  `json:"tags"`
}

type ListTagRes []*GetTagRes

type UpdateTagPreferenceReq struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}
