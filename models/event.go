package models

type Event struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`
	TodoId int64 `gorm:"not null" json:"todoId"`

	Name        string `gorm:"size:128" json:"name"`
	Description string `gorm:"size:256" json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      int32  `json:"sortId"`
}

type EventResponse struct {
	ModelResponse

	UserId string `json:"userId"`
	TodoId string `json:"todoId"`

	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      int32  `json:"sortId"`
}
