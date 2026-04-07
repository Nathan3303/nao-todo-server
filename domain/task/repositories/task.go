package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"

	"gorm.io/gorm"
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

	// BuildQueryTx 构建查询事务
	BuildQueryTx(ctx context.Context, query *valueobjects.QueryTask) (*gorm.DB, error)

	// ListWithQueryTx 获取任务列表（使用事务）
	ListWithQueryTx(
		ctx context.Context,
		tx *gorm.DB,
		pagination *valueobjects.Pagination,
	) ([]*entities.Task, *valueobjects.Pagination, error)
}
