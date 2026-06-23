package types

import "time"

type GetCheckItemRes struct {
	Id          string    `json:"id"`
	TaskId      string    `json:"taskId"`
	UserId      string    `json:"userId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDone      bool      `json:"isDone"`
	SortId      uint16    `json:"sortId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateCheckItemReq struct {
	TaskId      string `json:"taskId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateCheckItemRes struct {
	Id          string    `json:"id"`
	TaskId      string    `json:"taskId"`
	UserId      string    `json:"userId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDone      bool      `json:"isDone"`
	SortId      uint16    `json:"sortId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateCheckItemReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsDone      *bool   `json:"isDone"`
	SortId      *uint16 `json:"sortId"`
}

type ListCheckItemRes []*GetCheckItemRes

type ResortCheckItemsReq struct {
	OriginalId string `json:"originalId"`
	BoundId    string `json:"boundId"`
	Flag       int    `json:"flag"`
}

type ResortCheckItemsRes struct {
	OriginalSortId uint16 `json:"originalSortId"`
	BoundSortId    uint16 `json:"boundSortId"`
}

// BatchUpdateCheckItemReq 批量更新事件请求
type BatchUpdateCheckItemReq struct {
	Events []*struct {
		Id          string  `json:"id" binding:"required"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsDone      *bool   `json:"isDone"`
		SortId      *uint16 `json:"sortId"`
	} `json:"events" binding:"required,min=1"`
}

// BatchUpdateCheckItemRes 批量更新事件响应
type BatchUpdateCheckItemRes struct {
	UpdatedCount int64          `json:"updatedCount"`
	Events       []*GetCheckItemRes `json:"events"`
}
