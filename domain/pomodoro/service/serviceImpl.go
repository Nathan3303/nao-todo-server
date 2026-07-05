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

// GetById 获取 PomodoroRecord 详情
func (d *PomodoroDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	id int64,
) (*entities.PomodoroRecord, error) {
	return d.pomodoroRecordRepo.GetById(ctx, userId, id)
}

// List 获取 PomodoroRecord 列表
func (d *PomodoroDomainImpl) List(
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
) ([]*entities.PomodoroRecord, int64, error) {
	return d.pomodoroRecordRepo.List(
		ctx,
		userId,
		sessionId,
		startTime,
		endTime,
		taskId,
		taskName,
		pomodoroType,
		page,
		limit,
		sort,
	)
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

// GetPomodoroById 获取常用番茄工作详情
func (d *PomodoroDomainImpl) GetPomodoroById(
	ctx context.Context,
	userId int64,
	id int64,
) (*entities.Pomodoro, error) {
	return d.pomodoroRepo.GetById(ctx, userId, id)
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

// DeletePomodoro 删除常用番茄工作（软删除）
func (d *PomodoroDomainImpl) DeletePomodoro(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	return d.pomodoroRepo.Delete(ctx, userId, id)
}

// ArchivePomodoro 归档常用番茄工作
func (d *PomodoroDomainImpl) ArchivePomodoro(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	return d.pomodoroRepo.Archive(ctx, userId, id)
}

// UnarchivePomodoro 取消归档常用番茄工作
func (d *PomodoroDomainImpl) UnarchivePomodoro(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	return d.pomodoroRepo.Unarchive(ctx, userId, id)
}

// ListPomodoro 获取常用番茄工作列表
func (d *PomodoroDomainImpl) ListPomodoro(
	ctx context.Context,
	userId int64,
	pomodoroType uint8,
	name string,
	isArchived bool,
	page int,
	limit int,
	sort string,
) ([]*entities.Pomodoro, int64, error) {
	return d.pomodoroRepo.List(
		ctx,
		userId,
		pomodoroType,
		name,
		isArchived,
		page,
		limit,
		sort,
	)
}
