package task

import (
	"context"
	"errors"
	"naotodoserver/domain/task/service"
	"naotodoserver/domain/task/valueobjects"
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

// GetTaskById 获取单个任务信息
// @param ctx 上下文
// @param taskId 任务 ID
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) GetTaskById(
	ctx context.Context,
	taskId string,
) (*types.GetTaskRes, error) {
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
	res := TaskEntityToGetRes(taskEntity)
	return res, nil
}

// CreateTask 创建任务
// @param ctx 上下文
// @param req 创建任务请求
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) CreateTask(
	ctx context.Context,
	req *types.CreateTaskReq,
) (*types.GetTaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 请求体转换值对象
	createTaskValueObject, err := CreateTaskReqToValueObject(userId, req)
	if err != nil {
		return nil, err
	}
	// 3. 调用领域层创建任务
	taskEntity, err := taskApp.taskDomain.Create(ctx, userId, createTaskValueObject)
	if err != nil {
		return nil, err
	}
	// 4. 转换为响应对象
	res := TaskEntityToGetRes(taskEntity)
	return res, nil
}

// UpdateTask 更新任务
// @param ctx 上下文
// @param taskId 任务 ID
// @param req 更新任务请求
// @return 更新任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) UpdateTask(
	ctx context.Context,
	taskId string,
	req *types.UpdateTaskReq,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskIdInt64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return errors.New("待办任务 ID 无效")
	}
	// 3. 请求体转换值对象
	updateTaskValueObject, err := UpdateTaskReqToValueObject(userId, req)
	if err != nil {
		return err
	}
	// 4. 调用领域层更新任务
	err = taskApp.taskDomain.Update(ctx, userId, taskIdInt64, updateTaskValueObject)
	if err != nil {
		return err
	}
	// 5. 转换为响应对象
	return nil
}

// DeleteTask 删除任务
// @param ctx 上下文
// @param taskId 任务 ID
// @return 删除任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) DeleteTask(
	ctx context.Context,
	taskId string,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return errors.New("待办任务 ID 无效")
	}
	// 3. 调用领域层删除任务
	return taskApp.taskDomain.Delete(ctx, userId, taskId64)
}

// RestoreTask 恢复任务
// @param ctx 上下文
// @param taskId 任务 ID
// @return 恢复任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) RestoreTask(
	ctx context.Context,
	taskId string,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return errors.New("待办任务 ID 无效")
	}
	// 3. 调用领域层恢复任务
	return taskApp.taskDomain.Restore(ctx, userId, taskId64)
}

// CopyTask 复制任务
// @param ctx 上下文
// @param taskId 任务 ID
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) CopyTask(
	ctx context.Context,
	taskId string,
) (*types.GetTaskRes, error) {
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
	// 3. 调用领域层复制任务
	taskEntity, err := taskApp.taskDomain.Copy(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	// 4. 转换为响应对象
	res := TaskEntityToGetRes(taskEntity)
	return res, nil
}

// ListTask 获取任务列表
// @param ctx 上下文
// @param req 获取任务列表请求
// @return 任务列表响应
// @return error 错误信息
func (taskApp *TaskAppImpl) ListTask(
	ctx context.Context,
	req *types.ListTaskReq,
) (types.ListTaskRes, *types.Pagination, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, nil, errors.New("用户 ID 无效")
	}
	// 2. 请求体转换值对象
	queryTaskValueObject, err := ListTaskReqToQueryTaskValueObject(userId, req)
	if err != nil {
		return nil, nil, err
	}
	// 3. 分页值对象
	paginationValueObject := valueobjects.NewPagination(0, req.Page, req.Limit)
	// 4. 调用领域层获取任务列表
	taskEntities, paginationValueObject, err := taskApp.taskDomain.List(
		ctx,
		userId,
		queryTaskValueObject,
		paginationValueObject,
	)
	if err != nil {
		return nil, nil, err
	}
	// 5. 转换为响应对象
	tasks := TaskEntitiesToGetReses(taskEntities)
	paginationRes := PaginationValueObjectToRes(paginationValueObject)
	return tasks, paginationRes, nil
}
