package task

import (
	"context"
	"naotodoserver/domain/task/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type TaskApp interface {
	GetTaskById(ctx context.Context, taskId string) (*types.TaskRes, error)
	CreateTask(ctx context.Context, req *types.CreateTaskReq) (*types.TaskRes, error)
	UpdateTask(
		ctx context.Context,
		taskId string,
		req *types.UpdateTaskReq,
	) (*types.UpdateTaskRes, error)
	DeleteTask(ctx context.Context, taskId string) (*types.DeleteTaskRes, error)
	RestoreTask(ctx context.Context, taskId string) (*types.RestoreTaskRes, error)
	ListTask(
		ctx context.Context,
		req *types.ListTaskReq,
	) (types.ListTaskRes, *types.Pagination, error)
}

type TaskAppImpl struct {
	taskDomain service.TaskDomain
}

var (
	App  TaskApp
	once sync.Once
)
