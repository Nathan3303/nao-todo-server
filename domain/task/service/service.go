package service

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
)

// TaskDomain 任务领域服务接口
type TaskDomain interface {
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
	// @param createTaskValueObject 创建任务值对象
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
	// @param updateTaskValueObject 更新任务值对象
	// @return error 错误信息
	Update(
		ctx context.Context,
		userId int64,
		taskId int64,
		updateTaskValueObject *valueobjects.UpdateTask,
	) error

	// Delete 删除任务
	// @param ctx 上下文
	// @param userId 用户ID
	// @param taskId 任务ID
	// @return error 错误信息
	Delete(ctx context.Context, userId int64, taskId int64) error

	// Restore 恢复任务
	// @param ctx 上下文
	// @param userId 用户ID
	// @param taskId 任务ID
	// @return error 错误信息
	Restore(ctx context.Context, userId int64, taskId int64) error

	// Copy 复制任务
	// @param ctx 上下文
	// @param userId 用户ID
	// @param taskId 任务ID
	// @return 任务实体
	// @return error 错误信息
	Copy(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)

	// List 获取任务列表
	// @param ctx 上下文
	// @param userId 用户ID
	// @param query 任务查询参数
	// @param pagination 分页参数
	// @return 任务实体列表
	// @return pagination 分页信息
	// @return error 错误信息
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
	// @param durationMinutes 延迟分钟数
	// @return 新提醒时间
	// @return error 错误
	Snooze(ctx context.Context, userId int64, taskId int64, durationMinutes int) (string, error)
}

// TaskDomainImpl 任务领域服务实现
type TaskDomainImpl struct {
	taskRepo repositories.Task
}
