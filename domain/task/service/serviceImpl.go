package service

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
)

func NewTaskDomain(taskRepo repositories.Task) TaskDomain {
	return &TaskDomainImpl{taskRepo: taskRepo}
}

// GetById 获取单个任务信息
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return 任务实体
// @return error 错误信息
func (taskDomain *TaskDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	return taskDomain.taskRepo.GetById(ctx, userId, taskId)
}

// Create 创建任务
// @param ctx 上下文
// @param userId 用户ID
// @param createEntity 任务实体
// @return 任务实体
// @return error 错误信息
func (taskDomain *TaskDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createTaskValueObject *valueobjects.CreateTask,
) (*entities.Task, error) {
	return taskDomain.taskRepo.Create(ctx, userId, createTaskValueObject)
}

// Update 更新任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @param updateTaskValueObject 更新任务值对象
// @return error 错误
func (taskDomain *TaskDomainImpl) Update(
	ctx context.Context,
	userId int64,
	taskId int64,
	updateTaskValueObject *valueobjects.UpdateTask,
) error {
	return taskDomain.taskRepo.Update(ctx, userId, taskId, updateTaskValueObject)
}

// Delete 删除任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskDomain *TaskDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	taskId int64,
) error {
	return taskDomain.taskRepo.Delete(ctx, userId, taskId)
}

// Restore 恢复任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskDomain *TaskDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	taskId int64,
) error {
	return taskDomain.taskRepo.Restore(ctx, userId, taskId)
}

// List 获取任务列表
// @param ctx 上下文
// @param userId 用户ID
// @param query 查询参数
// @param pagination 分页参数
// @return 任务实体列表
// @return error 错误
func (taskDomain *TaskDomainImpl) List(
	ctx context.Context,
	userId int64,
	query *valueobjects.QueryTask,
	pagination *valueobjects.Pagination,
) ([]*entities.Task, *valueobjects.Pagination, error) {
	// 1. 补充 query
	query.UserId = userId
	// 2. 构建查询句柄
	_, err := taskDomain.taskRepo.BuildQueryTx(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	pagination.Page = query.Page
	pagination.Limit = query.Limit
	// 3. 查询任务列表
	taskEntities, pagination, err := taskDomain.taskRepo.List(ctx, userId, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	return taskEntities, pagination, nil
}
