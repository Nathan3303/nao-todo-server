package task

import (
	"context"
	"naotodoserver/application/task/dto"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/service"
)

// TaskApp 任务应用接口
type TaskApp interface {
	// --- Task ---
	GetTaskById(ctx context.Context, taskId string, includeDeleted bool) (*dto.GetTaskRes, error)
	CreateTask(ctx context.Context, req *dto.CreateTaskReq) (*dto.GetTaskRes, error)
	UpdateTask(ctx context.Context, taskId string, req *dto.UpdateTaskReq) error
	DeleteTask(ctx context.Context, taskId string) error
	RestoreTask(ctx context.Context, taskId string) error
	CopyTask(ctx context.Context, taskId string) (*dto.GetTaskRes, error)
	ListTask(
		ctx context.Context, req *dto.ListTaskReq,
	) (dto.ListTaskRes, *dto.Pagination, error)
	SnoozeTask(
		ctx context.Context, taskId string, req *dto.SnoozeTaskReq,
	) (*dto.SnoozeTaskRes, error)
	ProcessReminders(ctx context.Context) error
}

// TaskAppImpl 任务应用实现
type TaskAppImpl struct {
	taskDomain    service.TaskDomain
	taskRepo      repositories.Task
	checkItemRepo repositories.TaskCheckItem
	commentRepo   repositories.TaskComment
}
