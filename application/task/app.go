package task

import (
	"context"
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
	CreateTask(ctx context.Context, userId int64, req *dto.CreateTaskReq) (*dto.GetTaskRes, error)
	UpdateTask(ctx context.Context, userId int64, taskId string, req *dto.UpdateTaskReq) error
	DeleteTask(ctx context.Context, userId int64, taskId string) error
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
	taskDomain    service.TaskDomain
	taskRepo      repositories.Task
	checkItemRepo repositories.TaskCheckItem
	commentRepo   repositories.TaskComment
	publisher     domaintypes.NotificationPublisher
}
