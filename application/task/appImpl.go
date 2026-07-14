package task

import (
	"context"
	"fmt"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/application/idutil"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/service"
	"naotodoserver/domain/task/valueobjects"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/sse"
	"naotodoserver/interfaces/types"
	"time"
)

// NewTaskApp 创建任务应用层实例
func NewTaskApp(
	taskDomain service.TaskDomain,
	taskRepo repositories.Task,
	checkItemRepo repositories.TaskCheckItem,
	commentRepo repositories.TaskComment,
) *TaskAppImpl {
	impl := &TaskAppImpl{
		taskDomain:    taskDomain,
		taskRepo:      taskRepo,
		checkItemRepo: checkItemRepo,
		commentRepo:   commentRepo,
	}
	return impl
}

// GetTaskById 获取单个任务信息
// @param ctx 上下文
// @param taskId 任务 ID
// @param includeDeleted 为 true 时可查询到已软删除的任务
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) GetTaskById(
	ctx context.Context,
	taskId string,
	includeDeleted bool,
) (*types.GetTaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 转换待办任务 ID
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	// 3. 调用仓库获取待办任务信息
	taskEntity, err := taskApp.taskRepo.GetById(ctx, userId, taskId64, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("GetTaskById: %w", err)
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
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 请求体转换值对象
	createTaskValueObject, err := CreateTaskReqToValueObject(userId, req)
	if err != nil {
		return nil, err
	}
	taskEntity, err := taskApp.taskDomain.CreateTask(ctx, userId, createTaskValueObject)
	if err != nil {
		return nil, fmt.Errorf("CreateTask: %w", err)
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
		return domerr.ErrInvalidUserID
	}
	// 2. 转换待办任务 ID
	taskIdInt64, err := idutil.ParseID(taskId)
	if err != nil {
		return domerr.ErrInvalidTaskID
	}
	// 3. 请求体转换值对象
	updateTaskValueObject, err := UpdateTaskReqToValueObject(userId, req)
	if err != nil {
		return err
	}
	if err := taskApp.taskRepo.Update(ctx, userId, taskIdInt64, updateTaskValueObject); err != nil {
		return fmt.Errorf("UpdateTask: %w", err)
	}
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
		return domerr.ErrInvalidUserID
	}
	// 2. 转换待办任务 ID
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return domerr.ErrInvalidTaskID
	}
	if err := taskApp.taskRepo.Delete(ctx, userId, taskId64); err != nil {
		return fmt.Errorf("DeleteTask: %w", err)
	}
	return nil
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
		return domerr.ErrInvalidUserID
	}
	// 2. 转换待办任务 ID
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return domerr.ErrInvalidTaskID
	}
	if err := taskApp.taskRepo.Restore(ctx, userId, taskId64); err != nil {
		return fmt.Errorf("RestoreTask: %w", err)
	}
	return nil
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
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 转换待办任务 ID
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	// 3. 调用领域层复制任务
	taskEntity, err := taskApp.taskDomain.Copy(ctx, userId, taskId64)
	if err != nil {
		return nil, fmt.Errorf("CopyTask: %w", err)
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
		return nil, nil, domerr.ErrInvalidUserID
	}
	// 2. 请求体转换值对象
	queryTaskValueObject, err := ListTaskReqToQueryTaskValueObject(userId, req)
	if err != nil {
		return nil, nil, fmt.Errorf("ListTask req: %w", err)
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
		return nil, nil, fmt.Errorf("ListTask: %w", err)
	}
	// 5. 转换为响应对象
	tasks := TaskEntitiesToGetReses(taskEntities)
	paginationRes := PaginationValueObjectToRes(paginationValueObject)
	return tasks, paginationRes, nil
}

// SnoozeTask 稍后提醒
// @param ctx 上下文
// @param taskId 任务ID
// @param req 稍后提醒请求
// @return 稍后提醒响应
// @return error 错误
func (taskApp *TaskAppImpl) SnoozeTask(
	ctx context.Context,
	taskId string,
	req *types.SnoozeTaskReq,
) (*types.SnoozeTaskRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 2. 转换任务 ID
	taskIdInt64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	// 3. 调用领域层设置稍后提醒
	newRemindAt, err := taskApp.taskDomain.Snooze(
		ctx,
		userId,
		taskIdInt64,
		req.DurationMinutes,
	)
	if err != nil {
		return nil, fmt.Errorf("SnoozeTask: %w", err)
	}
	// 4. 返回结果
	return &types.SnoozeTaskRes{
		RemindAt: newRemindAt,
	}, nil
}

// ProcessReminders 处理所有到期提醒（供定时任务调用）
// @param ctx 上下文
// @return error 错误信息
func (taskApp *TaskAppImpl) ProcessReminders(ctx context.Context) error {
	tasks, err := taskApp.taskDomain.ProcessReminders(ctx)
	if err != nil {
		return fmt.Errorf("ProcessReminders: %w", err)
	}
	hub := sse.GetHub()
	for _, task := range tasks {
		hub.Publish(int64(task.UserId), sse.ReminderEvent{
			Type:        "REMINDER",
			TaskId:      idutil.FormatID(task.Id),
			TaskName:    task.Name,
			Description: task.Description,
			RemindAt:    task.RemindAt.ToString(time.RFC3339),
		})
	}
	return nil
}

// --- TaskCheckItem ---

// GetTaskCheckItemById 获取检查事项详情
// @param ctx 上下文
// @param itemId 检查事项 ID
// @return 检查事项响应
// @return error 错误信息
func (impl *TaskAppImpl) GetTaskCheckItemById(
	ctx context.Context,
	itemId string,
) (*types.GetTaskCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	id64, err := idutil.ParseID(itemId)
	if err != nil {
		return nil, domerr.ErrInvalidItemID
	}
	e, err := impl.checkItemRepo.GetCheckItemById(ctx, userId, id64)
	if err != nil {
		return nil, fmt.Errorf("GetCheckItemById: %w", err)
	}
	return TaskCheckItemEntityToGetRes(e), nil
}

// CreateTaskCheckItem 创建检查事项
// @param ctx 上下文
// @param req 创建检查事项请求
// @return 创建检查事项响应
// @return error 错误信息
func (impl *TaskAppImpl) CreateTaskCheckItem(
	ctx context.Context,
	req *types.CreateTaskCheckItemReq,
) (*types.CreateTaskCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	vo, err := CreateTaskCheckItemReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	e, err := impl.taskDomain.CreateCheckItem(ctx, userId, vo)
	if err != nil {
		return nil, fmt.Errorf("CreateCheckItem: %w", err)
	}
	return TaskCheckItemEntityToCreateRes(e), nil
}

// UpdateTaskCheckItem 更新检查事项
// @param ctx 上下文
// @param itemId 检查事项 ID
// @param req 更新检查事项请求
// @return error 错误信息
func (impl *TaskAppImpl) UpdateTaskCheckItem(
	ctx context.Context,
	itemId string,
	req *types.UpdateTaskCheckItemReq,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	id64, err := idutil.ParseID(itemId)
	if err != nil {
		return domerr.ErrInvalidItemID
	}
	vo, err := UpdateTaskCheckItemReqToVO(req)
	if err != nil {
		return err
	}
	if err := impl.checkItemRepo.UpdateCheckItem(ctx, userId, id64, vo); err != nil {
		return fmt.Errorf("UpdateCheckItem: %w", err)
	}
	return nil
}

// DeleteTaskCheckItem 删除检查事项
// @param ctx 上下文
// @param itemId 检查事项 ID
// @return error 错误信息
func (impl *TaskAppImpl) DeleteTaskCheckItem(
	ctx context.Context,
	itemId string,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	id64, err := idutil.ParseID(itemId)
	if err != nil {
		return domerr.ErrInvalidItemID
	}
	if err := impl.checkItemRepo.DeleteCheckItem(ctx, userId, id64); err != nil {
		return fmt.Errorf("DeleteCheckItem: %w", err)
	}
	return nil
}

// ListTaskCheckItems 获取检查事项列表
// @param ctx 上下文
// @param taskId 待办任务 ID
// @return 检查事项列表响应
// @return error 错误信息
func (impl *TaskAppImpl) ListTaskCheckItems(
	ctx context.Context,
	taskId string,
) (types.ListTaskCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	items, err := impl.checkItemRepo.ListCheckItems(ctx, userId, taskId64)
	if err != nil {
		return nil, fmt.Errorf("ListCheckItems: %w", err)
	}
	return TaskCheckItemEntitiesToReses(items), nil
}

// BatchUpdateTaskCheckItems 批量更新检查事项
// @param ctx 上下文
// @param req 批量更新检查事项请求
// @return 批量更新检查事项响应
// @return error 错误信息
func (impl *TaskAppImpl) BatchUpdateTaskCheckItems(
	ctx context.Context,
	req *types.BatchUpdateTaskCheckItemReq,
) (*types.BatchUpdateTaskCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	vos, err := BatchUpdateTaskCheckItemReqToVOs(req)
	if err != nil {
		return nil, err
	}
	items, err := impl.checkItemRepo.BatchUpdateCheckItems(ctx, userId, vos)
	if err != nil {
		return nil, fmt.Errorf("BatchUpdateCheckItems: %w", err)
	}
	resList := TaskCheckItemEntitiesToReses(items)
	return &types.BatchUpdateTaskCheckItemRes{
		UpdatedCount: int64(len(resList)),
		Events:       resList,
	}, nil
}

// --- Comment ---

// GetTaskCommentById 获取评论详情
// @param ctx 上下文
// @param commentId 评论 ID
// @return 评论响应
// @return error 错误信息
func (impl *TaskAppImpl) GetTaskCommentById(
	ctx context.Context,
	commentId string,
) (*types.TaskCommentRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	id64, err := idutil.ParseID(commentId)
	if err != nil {
		return nil, domerr.ErrInvalidCommentID
	}
	e, err := impl.commentRepo.GetCommentById(ctx, userId, id64)
	if err != nil {
		return nil, fmt.Errorf("GetCommentById: %w", err)
	}
	return TaskCommentEntityToRes(e), nil
}

// CreateTaskComment 创建评论
// @param ctx 上下文
// @param req 创建评论请求
// @return 创建评论响应
// @return error 错误信息
func (impl *TaskAppImpl) CreateTaskComment(
	ctx context.Context,
	req *types.CreateTaskCommentReq,
) (*types.TaskCommentRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	vo, err := CreateTaskCommentReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	e, err := impl.commentRepo.CreateComment(ctx, userId, vo)
	if err != nil {
		return nil, fmt.Errorf("CreateComment: %w", err)
	}
	return TaskCommentEntityToRes(e), nil
}

// UpdateTaskComment 更新评论
// @param ctx 上下文
// @param commentId 评论 ID
// @param req 更新评论请求
// @return error 错误信息
func (impl *TaskAppImpl) UpdateTaskComment(
	ctx context.Context,
	commentId string,
	req *types.UpdateTaskCommentReq,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	id64, err := idutil.ParseID(commentId)
	if err != nil {
		return domerr.ErrInvalidCommentID
	}
	vo, err := UpdateTaskCommentReqToVO(req)
	if err != nil {
		return err
	}
	if err := impl.commentRepo.UpdateComment(ctx, userId, id64, vo); err != nil {
		return fmt.Errorf("UpdateComment: %w", err)
	}
	return nil
}

// DeleteTaskComment 删除评论
// @param ctx 上下文
// @param commentId 评论 ID
// @return error 错误信息
func (impl *TaskAppImpl) DeleteTaskComment(
	ctx context.Context,
	commentId string,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	id64, err := idutil.ParseID(commentId)
	if err != nil {
		return domerr.ErrInvalidCommentID
	}
	if err := impl.commentRepo.DeleteComment(ctx, userId, id64); err != nil {
		return fmt.Errorf("DeleteComment: %w", err)
	}
	return nil
}

// ListTaskComments 获取评论列表
// @param ctx 上下文
// @param taskId 待办任务 ID
// @return 评论列表响应
// @return error 错误信息
func (impl *TaskAppImpl) ListTaskComments(
	ctx context.Context,
	taskId string,
) ([]*types.TaskCommentRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	entities, err := impl.commentRepo.ListComments(ctx, userId, taskId64)
	if err != nil {
		return nil, fmt.Errorf("ListComments: %w", err)
	}
	return TaskCommentEntitiesToListRes(entities), nil
}

// SyncTaskCommentUserProfile 同步评论用户信息
// @param ctx 上下文
// @param userId 用户 ID
// @param nickname 昵称
// @param avatar 头像
// @return error 错误信息
func (impl *TaskAppImpl) SyncTaskCommentUserProfile(
	ctx context.Context,
	userId int64,
	nickname, avatar string,
) error {
	return impl.commentRepo.SyncCommentUserProfile(ctx, userId, nickname, avatar)
}
