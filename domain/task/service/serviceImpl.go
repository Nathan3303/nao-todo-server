package service

import (
	"context"
	"errors"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/vo"
)

func NewTaskDomain(taskRepo repositories.Task) TaskDomain {
	return &TaskDomainImpl{taskRepo: taskRepo}
}

/*
 * Get task by id
 * 获取单个任务信息
 */
func (taskDomain *TaskDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	return taskDomain.taskRepo.GetById(ctx, &entities.Task{UserId: userId, Id: taskId})
}

/*
 * Create task
 * 创建任务
 */
func (taskDomain *TaskDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEntity *entities.Task,
) (*entities.Task, error) {
	if !createEntity.IsEndAtValid() {
		return nil, errors.New("时间参数无效 - 结束时间必须晚于开始时间")
	}
	createEntity.UserId = userId
	return taskDomain.taskRepo.Create(ctx, createEntity)
}

/*
 * Update task
 * 更新任务
 */
func (taskDomain *TaskDomainImpl) Update(
	ctx context.Context,
	userId int64,
	taskId int64,
	updateEntity *entities.Task,
) error {
	// err := updateEntity.IsDatesValid()
	// if err != nil {
	// 	return err
	// }
	whereEntity := &entities.Task{UserId: userId, Id: taskId}
	return taskDomain.taskRepo.Update(ctx, whereEntity, updateEntity)
}

/*
 * Delete task
 * 删除任务
 */
func (taskDomain *TaskDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	taskId int64,
) error {
	whereEntity := &entities.Task{UserId: userId, Id: taskId}
	return taskDomain.taskRepo.Delete(ctx, whereEntity)
}

/*
 * Restore task
 * 恢复任务
 */
func (taskDomain *TaskDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	taskId int64,
) error {
	whereEntity := &entities.Task{UserId: userId, Id: taskId}
	return taskDomain.taskRepo.Restore(ctx, whereEntity)
}

/*
 * List task
 * 获取任务列表
 */
func (taskDomain *TaskDomainImpl) List(
	ctx context.Context,
	userId int64,
	query *vo.TaskQuery,
	pagination *vo.Pagination,
) ([]*entities.Task, *vo.Pagination, error) {
	// 1. 补充 query
	query.UserId = userId
	// 2. 构建查询句柄
	tx, err := taskDomain.taskRepo.BuildQueryTx(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	pagination.Page = query.Page
	pagination.Limit = query.Limit
	// 3. 查询任务列表
	taskEntities, pagination, err := taskDomain.taskRepo.ListWithQueryTx(ctx, tx, pagination)
	if err != nil {
		return nil, nil, err
	}
	return taskEntities, pagination, nil
}
