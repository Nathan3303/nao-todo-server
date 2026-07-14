package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
)

// Task 任务仓库接口
type Task interface {
	// GetById 获取单个任务信息
	// includeDeleted 为 true 时可查询到已软删除的任务
	GetById(
		ctx context.Context,
		userId int64,
		taskId int64,
		includeDeleted bool,
	) (*entities.Task, error)

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

	// GetMaxSortId 获取任务最大排序 ID
	GetMaxSortId(ctx context.Context, userId int64) uint16

	// GetDueReminders 获取到期提醒任务
	GetDueReminders(ctx context.Context) ([]*entities.Task, error)

	// ClearRemindRepeat 清除提醒重复规则
	ClearRemindRepeat(ctx context.Context, taskId int64) error

	// UpdateRemindAt 更新提醒时间
	UpdateRemindAt(ctx context.Context, taskId int64, remindAt string) error
}
