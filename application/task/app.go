package task

import (
	"context"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/service"
	"naotodoserver/interfaces/types"
)

// TaskApp 任务应用接口
type TaskApp interface {
	// --- Task ---
	GetTaskById(ctx context.Context, taskId string, includeDeleted bool) (*types.GetTaskRes, error)
	CreateTask(ctx context.Context, req *types.CreateTaskReq) (*types.GetTaskRes, error)
	UpdateTask(ctx context.Context, taskId string, req *types.UpdateTaskReq) error
	DeleteTask(ctx context.Context, taskId string) error
	RestoreTask(ctx context.Context, taskId string) error
	CopyTask(ctx context.Context, taskId string) (*types.GetTaskRes, error)
	ListTask(
		ctx context.Context, req *types.ListTaskReq,
	) (types.ListTaskRes, *types.Pagination, error)
	SnoozeTask(
		ctx context.Context, taskId string, req *types.SnoozeTaskReq,
	) (*types.SnoozeTaskRes, error)
	ProcessReminders(ctx context.Context) error
}

// TaskAppImpl 任务应用实现
type TaskAppImpl struct {
	taskDomain service.TaskDomain
	taskRepo   repositories.Task
}
