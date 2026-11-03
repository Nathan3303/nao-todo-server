package task

import (
	"context"
	"time"

	"naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
)

// TaskCheckItemApp 任务检查项应用接口
type TaskCheckItemApp interface {
	GetTaskCheckItemById(
		ctx context.Context, userId int64, checkItemId string,
	) (*dto.GetTaskCheckItemRes, error)
	CreateTaskCheckItem(
		ctx context.Context, userId int64, req *dto.CreateTaskCheckItemReq,
	) (*dto.CreateTaskCheckItemRes, domaintypes.UpsertResult, error)
	UpdateTaskCheckItem(
		ctx context.Context, userId int64, checkItemId string, req *dto.UpdateTaskCheckItemReq,
	) error
	// DeleteTaskCheckItem 删除检查项（含 OCC 裁决，T642 / A4）：返回库中当前版本与裁决
	DeleteTaskCheckItem(
		ctx context.Context, userId int64, checkItemId string, baseUpdatedAt time.Time,
	) (*dto.GetTaskCheckItemRes, domaintypes.DeleteDecision, error)
	// GetTaskCheckItemSyncRow 读取单行检查项的同步形状（含墓碑），供 serverCurrent 使用
	GetTaskCheckItemSyncRow(
		ctx context.Context, userId int64, checkItemId string,
	) (*dto.GetTaskCheckItemRes, error)
	ListTaskCheckItems(
		ctx context.Context, userId int64, taskId string,
	) (dto.ListTaskCheckItemRes, error)
	ListTaskCheckItemSync(
		ctx context.Context, userId int64, updatedAt, cursorId string, limit int,
	) (dto.ListTaskCheckItemRes, error)
	BatchUpdateTaskCheckItems(
		ctx context.Context, userId int64, req *dto.BatchUpdateTaskCheckItemReq,
	) (*dto.BatchUpdateTaskCheckItemRes, error)
}
