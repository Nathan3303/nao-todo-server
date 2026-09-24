package service

import (
	"context"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

// PomodoroDomain Pomodoro 任务服务接口
type PomodoroDomain interface {
	// --- Pomodoro ---

	// CreatePomodoro 创建常用番茄工作
	// 返回 UpsertResult：Outcome 供同步回执使用；Created 供附属逻辑判定
	CreatePomodoro(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoro,
	) (*entities.Pomodoro, domaintypes.UpsertResult, error)

	// UpdatePomodoro 更新常用番茄工作（PATCH 语义）
	UpdatePomodoro(
		ctx context.Context,
		userId int64,
		id int64,
		vo *valueobjects.UpdatePomodoro,
	) (*entities.Pomodoro, error)

	// ListPomodoroSync 增量同步常用番茄工作列表（包含软删墓碑，(updated_at, id) keyset 游标稳定排序）
	ListPomodoroSync(
		ctx context.Context,
		userId int64,
		cursor time.Time,
		cursorID int64,
		limit int,
	) ([]*entities.Pomodoro, error)

	// --- PomodoroRecord ---

	// Create 创建 PomodoroRecord
	// 返回 UpsertResult：Outcome 供同步回执使用；Created 供附属逻辑判定
	CreatePomodoroRecord(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoroRecord,
	) (*entities.PomodoroRecord, domaintypes.UpsertResult, error)

	// ListPomodoroRecordSync 增量同步番茄工作记录列表（包含软删墓碑，(updated_at, id) keyset 游标稳定排序）
	ListPomodoroRecordSync(
		ctx context.Context,
		userId int64,
		cursor time.Time,
		cursorID int64,
		limit int,
	) ([]*entities.PomodoroRecord, error)
}

// PomodoroDomainImpl Pomodoro 任务服务实现
type PomodoroDomainImpl struct {
	pomodoroRecordRepo repositories.PomodoroRecord
	pomodoroRepo       repositories.Pomodoro
}
