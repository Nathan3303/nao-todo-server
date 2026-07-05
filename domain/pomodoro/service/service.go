package service

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// PomodoroDomain Pomodoro 任务服务接口
type PomodoroDomain interface {
	// --- PomodoroRecord ---

	// Create 创建 PomodoroRecord
	Create(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoroRecord,
	) (*entities.PomodoroRecord, error)

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
}

// PomodoroDomainImpl Pomodoro 任务服务实现
type PomodoroDomainImpl struct {
	pomodoroRecordRepo repositories.PomodoroRecord
	pomodoroRepo       repositories.Pomodoro
}
