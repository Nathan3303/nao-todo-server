package models

import (
	"database/sql"
	"time"
)

// Pomodoro 常用番茄工作模型
type Pomodoro struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_pomodoro_user_id"`

	// 番茄工作类型
	// 0: 番茄专注
	// 1: 正计时
	Type uint8 `gorm:"not null;default:0;type:tinyint(1)"`

	// 常用番茄工作名称
	Name string `gorm:"not null;size:256"`

	// 常用番茄工作描述
	Description string `gorm:"null;size:512"`

	// 常用番茄工作持续时间（秒）
	// 最大值为 3 小时，即 10800 秒
	// 最小值为 5 分钟，即 900 秒
	// ! 该属性在 Type 为 0 时才应被使用
	Duration uint16 `gorm:"not null;default:0;type:smallint"`

	// 归档时间
	ArchivedAt sql.NullTime `gorm:"null;index:idx_pomodoro_archived_at"`

	// --- 累加属性 ---

	// 番茄工作记录持续时间（秒）
	TotalDuration uint64 `gorm:"not null;default:0;type:bigint"`
}

// PomodoroRecord 番茄工作记录模型
type PomodoroRecord struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_pomodoro_user_id"`

	// 常用番茄工作 ID
	// 弱关联，可以为空
	PomodoroId int64 `gorm:"null;index:idx_pomodoro_pomodoro_id"`

	// 番茄工作会话 UUID
	SessionId string `gorm:"not null;size:36;index:idx_pomodoro_session_id"`

	// 番茄工作类型
	// 0: 番茄专注
	// 1: 正计时
	Type uint8 `gorm:"not null;default:0;type:tinyint(1)"`

	// 待办任务 ID
	TaskId int64 `gorm:"not null;index:idx_pomodoro_task_id"`

	// 待办任务名称
	TaskName string `gorm:"not null;size:256"`

	// 番茄工作描述
	Description string `gorm:"null;size:512"`

	// 番茄工作开始时间
	StartAt time.Time `gorm:"not null;index:idx_pomodoro_start_at"`

	// 番茄工作结束时间
	EndAt time.Time `gorm:"not null"`

	// 番茄工作持续时间（秒）
	// 最大值为 3 小时，即 10800 秒
	// 最小值为 5 分钟，即 900 秒
	Duration uint16 `gorm:"not null;default:0;type:smallint"`

	// 番茄工作笔记
	Note string `gorm:"null;type:text"`

	// 归档时间
	ArchivedAt sql.NullTime `gorm:"null;index:idx_pomodoro_archived_at"`
}
