package service

import (
	"context"
	"time"

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
	// 幂等创建：客户端指定 id 时走 upsert（LWW + create 冲突检测）
	entity, _, err := d.pomodoroRepo.Upsert(ctx, userId, vo)
	return entity, err
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

// --- PomodoroRecord ---

// CreatePomodoroRecord 创建 PomodoroRecord
func (d *PomodoroDomainImpl) CreatePomodoroRecord(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreatePomodoroRecord,
) (*entities.PomodoroRecord, error) {
	vo.UserId = userId
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	// 幂等创建：客户端指定 id 时走 upsert（LWW + create 冲突检测）
	entity, _, err := d.pomodoroRecordRepo.Upsert(ctx, userId, vo)
	return entity, err
}

// ListPomodoroSync 增量同步常用番茄工作列表
func (d *PomodoroDomainImpl) ListPomodoroSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.Pomodoro, error) {
	return d.pomodoroRepo.ListSync(ctx, userId, cursor, cursorID, limit)
}

// ListPomodoroRecordSync 增量同步番茄工作记录列表
func (d *PomodoroDomainImpl) ListPomodoroRecordSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.PomodoroRecord, error) {
	return d.pomodoroRecordRepo.ListSync(ctx, userId, cursor, cursorID, limit)
}
