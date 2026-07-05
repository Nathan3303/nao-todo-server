package service

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// NewPomodoroDomain 创建 Pomodoro 任务服务实现
func NewPomodoroDomain(
	pomodoroRecordRepo repositories.PomodoroRecord,
	pomodoroRepo repositories.Pomodoro,
) PomodoroDomain {
	return &PomodoroDomainImpl{
		pomodoroRecordRepo: pomodoroRecordRepo,
		pomodoroRepo:       pomodoroRepo,
	}
}

// --- PomodoroRecord ---

// Create 创建 PomodoroRecord
func (d *PomodoroDomainImpl) Create(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreatePomodoroRecord,
) (*entities.PomodoroRecord, error) {
	vo.UserId = userId
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return d.pomodoroRecordRepo.Create(ctx, vo)
}

// --- Pomodoro ---

// CreatePomodoro 创建常用番茄工作
func (d *PomodoroDomainImpl) CreatePomodoro(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreatePomodoro,
) (*entities.Pomodoro, error) {
	vo.UserId = userId
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return d.pomodoroRepo.Create(ctx, vo)
}

// UpdatePomodoro 更新常用番茄工作（PATCH 语义）
func (d *PomodoroDomainImpl) UpdatePomodoro(
	ctx context.Context,
	userId int64,
	id int64,
	vo *valueobjects.UpdatePomodoro,
) (*entities.Pomodoro, error) {
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return d.pomodoroRepo.Update(ctx, userId, id, vo)
}


