package entities

import (
	"naotodoserver/domain/types"
	"time"
)

// Pomodoro 待办任务番茄工作实体
// 用于表示用户在系统中的待办任务番茄工作记录
// 包含用户 ID、会话 ID、类型、任务 ID、任务名称、描述、开始时间、结束时间、持续时间、备注等属性
type Pomodoro struct {
	types.EntityBase
	UserId      int64
	SessionId   string
	Type        uint8
	TaskId      int64
	TaskName    string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Duration    uint8
	Note        string
}
