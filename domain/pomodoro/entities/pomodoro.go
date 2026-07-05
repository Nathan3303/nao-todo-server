package entities

import "naotodoserver/domain/types"

// Pomodoro 专注工作记录实体
type Pomodoro struct {
	types.EntityBase
	UserId        int64
	Type          uint8
	Name          string
	Description   string
	Duration      uint16
	ArchivedAt    types.NullableTime
	TotalDuration uint64
}
