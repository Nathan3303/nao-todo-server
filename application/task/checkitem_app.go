package task

import (
	"context"
	"naotodoserver/application/task/dto"
)

// TaskCheckItemApp 任务检查项应用接口
type TaskCheckItemApp interface {
	GetTaskCheckItemById(
		ctx context.Context, checkItemId string,
	) (*dto.GetTaskCheckItemRes, error)
	CreateTaskCheckItem(
		ctx context.Context, req *dto.CreateTaskCheckItemReq,
	) (*dto.CreateTaskCheckItemRes, error)
	UpdateTaskCheckItem(
		ctx context.Context, checkItemId string, req *dto.UpdateTaskCheckItemReq,
	) error
	DeleteTaskCheckItem(ctx context.Context, checkItemId string) error
	ListTaskCheckItems(ctx context.Context, taskId string) (dto.ListTaskCheckItemRes, error)
	BatchUpdateTaskCheckItems(
		ctx context.Context, req *dto.BatchUpdateTaskCheckItemReq,
	) (*dto.BatchUpdateTaskCheckItemRes, error)
}
