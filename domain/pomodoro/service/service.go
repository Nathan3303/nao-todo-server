package service

import (
	"context"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// PomodoroDomain Pomodoro 任务服务接口
type PomodoroDomain interface {
	// --- Pomodoro ---

	// CreatePomodoro 创建常用番茄工作
	CreatePomodoro(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoro,
	) (*entities.Pomodoro, error)

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
	CreatePomodoroRecord(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoroRecord,
	) (*entities.PomodoroRecord, error)

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
