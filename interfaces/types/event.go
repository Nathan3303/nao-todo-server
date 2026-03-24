package types

type EventRes struct {
	Id          string `json:"id"`
	TaskId      string `json:"taskId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      uint32 `json:"sortId"`
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
	SortId      uint32 `json:"sortId"`
}

type UpdateEventRes struct {
	EventId string `json:"eventId"`
}

type DeleteEventRes struct {
	EventId string `json:"eventId"`
}

type ListEventRes []*EventRes

type ResortEventsReq struct {
	OriginalId string `json:"originalId"`
	BoundId    string `json:"boundId"`
	Flag       int    `json:"flag"`
}

type ResortEventsRes struct {
	OriginalSortId uint32 `json:"originalSortId"`
	BoundSortId    uint32 `json:"boundSortId"`
}
