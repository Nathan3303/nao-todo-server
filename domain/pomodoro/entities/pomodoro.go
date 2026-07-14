package entities

import "naotodoserver/domain/types"

// Pomodoro 专注工作记录实体
type Pomodoro struct {
	types.EntityBase
	UserId        types.UserID
	Type          PomodoroType
	Name          string
	Description   string
	Duration      uint16
	ArchivedAt    types.NullableTime
	TotalDuration uint64
}
