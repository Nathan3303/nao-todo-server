package task

import (
	"context"
	"naotodoserver/application/task/dto"
)

// TaskCheckItemApp 任务检查项应用接口
type TaskCheckItemApp interface {
	GetTaskCheckItemById(
		ctx context.Context, userId int64, checkItemId string,
	) (*dto.GetTaskCheckItemRes, error)
	CreateTaskCheckItem(
		ctx context.Context, userId int64, req *dto.CreateTaskCheckItemReq,
	) (*dto.CreateTaskCheckItemRes, error)
	UpdateTaskCheckItem(
		ctx context.Context, userId int64, checkItemId string, req *dto.UpdateTaskCheckItemReq,
	) error
	DeleteTaskCheckItem(ctx context.Context, userId int64, checkItemId string) error
	ListTaskCheckItems(
		ctx context.Context, userId int64, taskId string,
	) (dto.ListTaskCheckItemRes, error)
	BatchUpdateTaskCheckItems(
		ctx context.Context, userId int64, req *dto.BatchUpdateTaskCheckItemReq,
	) (*dto.BatchUpdateTaskCheckItemRes, error)
}
