package service

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// NewPomodoroDomain 创建 Pomodoro 任务服务实现
func NewPomodoroDomain(pomodoroRepo repositories.Pomodoro) PomodoroDomain {
	return &PomodoroDomainImpl{pomodoroRepo: pomodoroRepo}
}

// Create 创建 Pomodoro 任务
func (d *PomodoroDomainImpl) Create(
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

// GetById 获取 Pomodoro 任务详情
func (d *PomodoroDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	id int64,
) (*entities.Pomodoro, error) {
	return d.pomodoroRepo.GetById(ctx, userId, id)
}

// List 获取 Pomodoro 任务列表
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
) ([]*entities.Pomodoro, int64, error) {
	return d.pomodoroRepo.List(
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
