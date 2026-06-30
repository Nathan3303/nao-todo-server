package repositories

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// PomodoroRecord Pomodoro 任务记录接口
type PomodoroRecord interface {
	// Create 创建番茄工作记录
	Create(
		ctx context.Context,
		vo *valueobjects.CreatePomodoroRecord,
	) (*entities.PomodoroRecord, error)

	// GetById 根据 ID 获取番茄工作记录
	GetById(ctx context.Context, userId int64, id int64) (*entities.PomodoroRecord, error)

	// List 获取番茄工作记录列表
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
}
