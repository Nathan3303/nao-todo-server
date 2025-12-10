package types

type EventRes struct {
	Id          string `json:"id"`
	TaskId      string `json:"taskId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      int32  `json:"sortId"`
}

type CreateEventReq struct {
	TaskId      string `json:"taskId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateEventRes EventRes

type UpdateEventReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      int32  `json:"sortId"`
}

type UpdateEventRes struct {
	EventId string `json:"eventId"`
}

type DeleteEventRes struct {
	EventId string `json:"eventId"`
}

type ListEventRes []*EventRes
