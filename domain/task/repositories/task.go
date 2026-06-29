package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
)

// Task 任务仓库接口
type Task interface {
	// GetById 获取单个任务信息
	GetById(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)

	// Create 创建任务
	Create(
		ctx context.Context,
		userId int64,
		createTaskValueObject *valueobjects.CreateTask,
	) (*entities.Task, error)

	// Update 更新任务
	Update(
		ctx context.Context,
		userId int64,
		taskId int64,
		updateTaskValueObject *valueobjects.UpdateTask,
	) error

	// Delete 删除任务
	Delete(ctx context.Context, userId int64, taskId int64) error

	// Restore 恢复任务
	Restore(ctx context.Context, userId int64, taskId int64) error

	// List 获取任务列表
	List(
		ctx context.Context,
		userId int64,
		query *valueobjects.QueryTask,
		pagination *valueobjects.Pagination,
	) ([]*entities.Task, *valueobjects.Pagination, error)

	// Snooze 稍后提醒
	Snooze(ctx context.Context, userId int64, taskId int64, remindAt string) error

	// GetDueReminders 获取到期提醒任务
	GetDueReminders(ctx context.Context) ([]*entities.Task, error)

	// ClearRemindRepeat 清除提醒重复规则
	ClearRemindRepeat(ctx context.Context, taskId int64) error

	// UpdateRemindAt 更新提醒时间
	UpdateRemindAt(ctx context.Context, taskId int64, remindAt string) error

	// --- CheckItem ---

	// GetCheckItemById 获取单个任务检查项信息
	GetCheckItemById(
		ctx context.Context,
		userId int64,
		checkItemId int64,
	) (*entities.TaskCheckItem, error)

	// CreateCheckItem 创建任务检查项
	CreateCheckItem(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskCheckItem,
	) (*entities.TaskCheckItem, error)

	// UpdateCheckItem 更新任务检查项
	UpdateCheckItem(
		ctx context.Context,
		userId int64,
		checkItemId int64,
		vo *valueobjects.UpdateTaskCheckItem,
	) error

	// DeleteCheckItem 删除任务检查项
	DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error

	// ListCheckItems 获取任务检查项列表
	ListCheckItems(ctx context.Context, userId, taskId int64) ([]*entities.TaskCheckItem, error)

	// GetMaxCheckItemSortId 获取任务检查项最大排序 ID
	GetMaxCheckItemSortId(ctx context.Context, userId, taskId int64) uint16

	// BatchUpdateCheckItems 批量更新任务检查项
	BatchUpdateCheckItems(
		ctx context.Context,
		userId int64,
		vos []*valueobjects.BatchUpdateTaskCheckItem,
	) ([]*entities.TaskCheckItem, error)

	// --- Comment ---

	// GetCommentById 获取单个任务评论信息
	GetCommentById(ctx context.Context, userId, commentId int64) (*entities.TaskComment, error)

	// CreateComment 创建任务评论
	CreateComment(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskComment,
	) (*entities.TaskComment, error)

	// UpdateComment 更新任务评论
	UpdateComment(
		ctx context.Context,
		userId int64,
		commentId int64,
		vo *valueobjects.UpdateTaskComment,
	) error

	// DeleteComment 删除任务评论
	DeleteComment(ctx context.Context, userId, commentId int64) error

	// ListComments 获取任务评论列表
	ListComments(ctx context.Context, userId, taskId int64) ([]*entities.TaskComment, error)

	// SyncCommentUserProfile 同步任务评论用户信息
	SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}
