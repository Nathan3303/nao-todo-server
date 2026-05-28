package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
)

// Task 任务仓库接口
type Task interface {
	// GetById 获取单个任务信息
	// @param ctx 上下文
	// @param userId 用户ID
	// @param taskId 任务ID
	// @return 任务实体
	// @return error 错误信息
	GetById(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)

	// Create 创建任务
	// @param ctx 上下文
	// @param userId 用户ID
	// @param createTaskValueObject 任务创建值对象
	// @return 任务实体
	// @return error 错误信息
	Create(
		ctx context.Context,
		userId int64,
		createTaskValueObject *valueobjects.CreateTask,
	) (*entities.Task, error)

	// Update 更新任务
	// @param ctx 上下文
	// @param userId 用户ID
	// @param taskId 任务ID
	// @param updateEntity 任务实体
	// @return error 错误信息
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
	// @param ctx 上下文
	// @param userId 用户ID
	// @param taskId 任务ID
	// @param remindAt 新提醒时间
	// @return error 错误
	Snooze(ctx context.Context, userId int64, taskId int64, remindAt string) error

	// GetDueReminders 获取到期提醒任务
	// @param ctx 上下文
	// @return 任务实体列表
	// @return error 错误
	GetDueReminders(ctx context.Context) ([]*entities.Task, error)

	// ClearRemindRepeat 清除提醒重复规则
	// @param ctx 上下文
	// @param taskId 任务ID
	// @return error 错误
	ClearRemindRepeat(ctx context.Context, taskId int64) error

	// UpdateRemindAt 更新提醒时间
	// @param ctx 上下文
	// @param taskId 任务ID
	// @param remindAt 新提醒时间
	// @return error 错误
	UpdateRemindAt(ctx context.Context, taskId int64, remindAt string) error
}
