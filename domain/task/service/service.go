package service

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
)

// TaskDomain 任务域接口
type TaskDomain interface {
	// --- Task ---
	Copy(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)
	CreateTask(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTask,
	) (*entities.Task, error)
	List(
		ctx context.Context,
		userId int64,
		query *valueobjects.QueryTask,
		pagination *valueobjects.Pagination,
	) ([]*entities.Task, *valueobjects.Pagination, error)

	// --- CheckItem ---
	CreateCheckItem(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskCheckItem,
	) (*entities.TaskCheckItem, error)

	// --- Snooze ---
	Snooze(ctx context.Context, userId int64, taskId int64, durationMinutes int) (string, error)
	ProcessReminders(ctx context.Context) ([]*entities.Task, error)
}

// TaskDomainImpl 任务域实现
type TaskDomainImpl struct {
	taskRepo      repositories.Task
	checkItemRepo repositories.TaskCheckItem
}
