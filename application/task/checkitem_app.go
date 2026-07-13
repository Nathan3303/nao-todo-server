package task

import (
	"context"
	"naotodoserver/interfaces/types"
)

// TaskCheckItemApp 任务检查项应用接口
type TaskCheckItemApp interface {
	GetTaskCheckItemById(
		ctx context.Context, checkItemId string,
	) (*types.GetTaskCheckItemRes, error)
	CreateTaskCheckItem(
		ctx context.Context, req *types.CreateTaskCheckItemReq,
	) (*types.CreateTaskCheckItemRes, error)
	UpdateTaskCheckItem(
		ctx context.Context, checkItemId string, req *types.UpdateTaskCheckItemReq,
	) error
	DeleteTaskCheckItem(ctx context.Context, checkItemId string) error
	ListTaskCheckItems(ctx context.Context, taskId string) (types.ListTaskCheckItemRes, error)
	BatchUpdateTaskCheckItems(
		ctx context.Context, req *types.BatchUpdateTaskCheckItemReq,
	) (*types.BatchUpdateTaskCheckItemRes, error)
}
