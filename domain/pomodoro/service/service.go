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

	// GetById 获取 PomodoroRecord 详情
	GetById(ctx context.Context, userId int64, id int64) (*entities.PomodoroRecord, error)

	// List 获取 PomodoroRecord 列表
	List(
		ctx context.Context,
		userId int64,
		sessionId string,
		startTime string,
		endTime string,
		taskId int64,
		taskName string,
		pomodoroType uint8,
		page int,
		limit int,
		sort string,
	) ([]*entities.PomodoroRecord, int64, error)

	// --- Pomodoro ---

	// CreatePomodoro 创建常用番茄工作
	CreatePomodoro(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoro,
	) (*entities.Pomodoro, error)

	// GetPomodoroById 获取常用番茄工作详情
	GetPomodoroById(
		ctx context.Context,
		userId int64,
		id int64,
	) (*entities.Pomodoro, error)

	// UpdatePomodoro 更新常用番茄工作（PATCH 语义）
	UpdatePomodoro(
		ctx context.Context,
		userId int64,
		id int64,
		vo *valueobjects.UpdatePomodoro,
	) (*entities.Pomodoro, error)

	// DeletePomodoro 删除常用番茄工作（软删除）
	DeletePomodoro(ctx context.Context, userId int64, id int64) error

	// ArchivePomodoro 归档常用番茄工作
	ArchivePomodoro(ctx context.Context, userId int64, id int64) error

	// UnarchivePomodoro 取消归档常用番茄工作
	UnarchivePomodoro(ctx context.Context, userId int64, id int64) error

	// ListPomodoro 获取常用番茄工作列表
	ListPomodoro(
		ctx context.Context,
		userId int64,
		pomodoroType uint8,
		name string,
		isArchived bool,
		page int,
		limit int,
		sort string,
	) ([]*entities.Pomodoro, int64, error)
}

// PomodoroDomainImpl Pomodoro 任务服务实现
type PomodoroDomainImpl struct {
	pomodoroRecordRepo repositories.PomodoroRecord
	pomodoroRepo       repositories.Pomodoro
}
