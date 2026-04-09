package types

import "time"

type GetEventRes struct {
	Id          string    `json:"id"`
	TaskId      string    `json:"taskId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDone      bool      `json:"isDone"`
	SortId      uint16    `json:"sortId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateEventReq struct {
	TaskId      string `json:"taskId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateEventRes struct {
	Id          string    `json:"id"`
	TaskId      string    `json:"taskId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDone      bool      `json:"isDone"`
	SortId      uint16    `json:"sortId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateEventReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      uint16 `json:"sortId"`
}

type ListEventRes []*GetEventRes

type ResortEventsReq struct {
	OriginalId string `json:"originalId"`
	BoundId    string `json:"boundId"`
	Flag       int    `json:"flag"`
}

type ResortEventsRes struct {
	OriginalSortId uint16 `json:"originalSortId"`
	BoundSortId    uint16 `json:"boundSortId"`
}
