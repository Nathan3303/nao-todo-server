package service

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// PomodoroDomain Pomodoro 任务服务接口
type PomodoroDomain interface {
	Create(ctx context.Context, userId int64, vo *valueobjects.CreatePomodoro) (*entities.Pomodoro, error)
	GetById(ctx context.Context, userId int64, id int64) (*entities.Pomodoro, error)
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
	) ([]*entities.Pomodoro, int64, error)
}

// PomodoroDomainImpl Pomodoro 任务服务实现
type PomodoroDomainImpl struct {
	pomodoroRepo repositories.Pomodoro
}
