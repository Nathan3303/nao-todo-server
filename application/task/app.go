package task

import (
	"context"
	"time"

	"naotodoserver/application/task/dto"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/service"
	domaintypes "naotodoserver/domain/types"
)

// TaskApp 任务应用接口
type TaskApp interface {
	// --- Task ---
	GetTaskById(
		ctx context.Context, userId int64, taskId string, includeDeleted bool,
	) (*dto.GetTaskRes, error)
	CreateTask(
		ctx context.Context,
		userId int64,
		req *dto.CreateTaskReq,
	) (*dto.GetTaskRes, domaintypes.UpsertResult, error)
	UpdateTask(ctx context.Context, userId int64, taskId string, req *dto.UpdateTaskReq) error
	// DeleteTask 删除任务（含 OCC 裁决，T642 / A4）：baseUpdatedAt 为零值 = 未提供（回退现行「删除优先」）
	// 返回库中当前版本（sync 条目形状，含墓碑；行不存在时为 nil）与删除裁决
	DeleteTask(
		ctx context.Context, userId int64, taskId string, baseUpdatedAt time.Time,
	) (*dto.GetTaskRes, domaintypes.DeleteDecision, error)
	RestoreTask(ctx context.Context, userId int64, taskId string) error
	CopyTask(ctx context.Context, userId int64, taskId string) (*dto.GetTaskRes, error)
	ListTask(
		ctx context.Context, userId int64, req *dto.ListTaskReq,
	) (dto.ListTaskRes, *dto.Pagination, error)
	// ListTaskSync 增量同步任务列表（包含软删墓碑，updated_at 游标稳定排序分页）
	ListTaskSync(
		ctx context.Context, userId int64, req *dto.ListTaskReq,
	) (dto.ListTaskRes, error)
	SnoozeTask(
		ctx context.Context, userId int64, taskId string, req *dto.SnoozeTaskReq,
	) (*dto.SnoozeTaskRes, error)
	ProcessReminders(ctx context.Context) error
}

// TaskAppImpl 任务应用实现
type TaskAppImpl struct {
	taskDomain     service.TaskDomain
	taskRepo       repositories.Task
	checkItemRepo  repositories.TaskCheckItem
	commentRepo    repositories.TaskComment
	publisher      domaintypes.NotificationPublisher
	txManager      domaintypes.TxManager
	countPublisher domaintypes.CountEventPublisher
}
