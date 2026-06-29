package repositories

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// Pomodoro Pomodoro 任务接口
type Pomodoro interface {
	Create(ctx context.Context, vo *valueobjects.CreatePomodoro) (*entities.Pomodoro, error)
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
