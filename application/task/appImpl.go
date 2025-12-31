package task

import (
	"context"
	"errors"
	"naotodoserver/domain/task/service"
	"naotodoserver/domain/task/vo"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// RegistDomainImpl 注册任务领域层
func RegistDomainImpl(taskDomain service.TaskDomain) TaskApp {
	once.Do(func() {
		App = &TaskAppImpl{
			taskDomain: taskDomain,
		}
	})
	return App
}

/*
 * Get task by id
 * 获取单个任务信息
 */
func (taskApp *TaskAppImpl) GetTaskById(
	ctx context.Context,
	taskId string,
) (*types.TaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 无效")
	}
	// 3. 调用领域层获取待办任务信息
	taskEntity, err := taskApp.taskDomain.GetById(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	// 4. 转换为响应对象
	res := TaskEntity2Res(taskEntity)
	return res, nil
}

/*
 * Create task
 * 创建任务
 */
func (taskApp *TaskAppImpl) CreateTask(
	ctx context.Context,
	req *types.CreateTaskReq,
) (*types.TaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 判断清单 ID 是否为空 - 为空则使用用户 ID 作为清单 ID（收集箱）
	if req.ProjectId == "" {
		req.ProjectId = strconv.FormatInt(userId, 10)
	}
	// 2. 请求体转换实体
	createEntity := CreateTaskReq2Entity(req)
	// 3. 调用领域层创建任务
	taskEntity, err := taskApp.taskDomain.Create(ctx, userId, createEntity)
	if err != nil {
		return nil, err
	}
	// 4. 转换为响应对象
	res := TaskEntity2Res(taskEntity)
	return res, nil
}

/*
 * Update task
 * 创建任务
 */
func (taskApp *TaskAppImpl) UpdateTask(
	ctx context.Context,
	taskId string,
	req *types.UpdateTaskReq,
) (*types.UpdateTaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 无效")
	}
	// 3. 请求体转换实体
	updateEntity := UpdateTaskReq2Entity(req)
	// 4. 调用领域层更新任务
	err = taskApp.taskDomain.Update(ctx, userId, taskId64, updateEntity)
	if err != nil {
		return nil, err
	}
	// 5. 转换为响应对象
	return &types.UpdateTaskRes{TaskId: taskId}, nil
}

/*
 * Delete task
 * 删除任务
 */
func (taskApp *TaskAppImpl) DeleteTask(
	ctx context.Context,
	taskId string,
) (*types.DeleteTaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 无效")
	}
	// 3. 调用领域层删除任务
	err = taskApp.taskDomain.Delete(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	return &types.DeleteTaskRes{TaskId: taskId}, nil
}

/*
 * Restore task
 * 恢复任务
 */
func (taskApp *TaskAppImpl) RestoreTask(
	ctx context.Context,
	taskId string,
) (*types.RestoreTaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 无效")
	}
	// 3. 调用领域层恢复任务
	err = taskApp.taskDomain.Restore(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	return &types.RestoreTaskRes{TaskId: taskId}, nil
}

/*
 * List task
 * 获取任务列表
 */
func (taskApp *TaskAppImpl) ListTask(
	ctx context.Context,
	req *types.ListTaskReq,
) (types.ListTaskRes, *types.Pagination, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, nil, errors.New("用户 ID 无效")
	}
	// 2. 请求体转换实体
	query := ListTaskReq2QueryVO(req)
	// 3. 调用领域层获取任务列表
	pagination := &vo.Pagination{Page: req.Page, Limit: req.Limit}
	taskEntities, pagination, err := taskApp.taskDomain.List(ctx, userId, query, pagination)
	if err != nil {
		return nil, nil, err
	}
	// 4. 转换为响应对象
	tasks := TaskEntities2Reses(taskEntities)
	paginationRes := PaginationVO2Res(pagination)
	return tasks, paginationRes, nil
}
