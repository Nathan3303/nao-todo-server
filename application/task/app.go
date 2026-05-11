package task

import (
	"context"
	"naotodoserver/domain/task/service"
	"naotodoserver/interfaces/types"
	"sync"
)

// TaskApp 任务应用层
type TaskApp interface {
	// GetTaskById 获取单个任务信息
	GetTaskById(ctx context.Context, taskId string) (*types.GetTaskRes, error)

	// CreateTask 创建任务
	CreateTask(ctx context.Context, req *types.CreateTaskReq) (*types.GetTaskRes, error)

	// UpdateTask 更新任务
	UpdateTask(
		ctx context.Context,
		taskId string,
		req *types.UpdateTaskReq,
	) error

	// DeleteTask 删除任务
	DeleteTask(ctx context.Context, taskId string) error

	// RestoreTask 恢复任务
	RestoreTask(ctx context.Context, taskId string) error

	// CopyTask 复制任务
	CopyTask(ctx context.Context, taskId string) (*types.GetTaskRes, error)

	// ListTask 获取任务列表
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
