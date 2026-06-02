package task

import (
	"context"
	"errors"
	"naotodoserver/domain/task/service"
	"naotodoserver/domain/task/valueobjects"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/sse"
	"naotodoserver/interfaces/types"
	"strconv"
)

// NewTaskApp 创建任务应用层实例
func NewTaskApp(taskDomain service.TaskDomain) TaskApp {
	impl := &TaskAppImpl{
		taskDomain: taskDomain,
	}
	return impl
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
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换任务 ID
	taskIdInt64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("任务 ID 无效")
	}
	// 3. 调用领域层设置稍后提醒
	newRemindAt, err := taskApp.taskDomain.Snooze(ctx, userId, taskIdInt64, req.DurationMinutes)
	if err != nil {
		return nil, err
	}
	// 4. 返回结果
	return &types.SnoozeTaskRes{
		RemindAt: newRemindAt,
	}, nil
}

// ProcessReminders 处理所有到期提醒（供定时任务调用）
func (taskApp *TaskAppImpl) ProcessReminders(ctx context.Context) error {
	tasks, err := taskApp.taskDomain.ProcessReminders(ctx)
	if err != nil {
		return err
	}
	hub := sse.GetHub()
	for _, task := range tasks {
		hub.Publish(task.UserId, sse.ReminderEvent{
			Type:        "REMINDER",
			TaskId:      strconv.FormatInt(task.Id, 10),
			TaskName:    task.Name,
			Description: task.Description,
			RemindAt:    formatTimePtr(task.RemindAt),
		})
	}
	return nil
}

// === CheckItem ===

func (impl *TaskAppImpl) GetCheckItemById(ctx context.Context, itemId string) (*types.GetCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	id64, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		return nil, errors.New("检查事项 ID 格式错误")
	}
	e, err := impl.taskDomain.GetCheckItemById(ctx, userId, id64)
	if err != nil {
		return nil, err
	}
	return CheckItemEntityToGetRes(e), nil
}

func (impl *TaskAppImpl) CreateCheckItem(ctx context.Context, req *types.CreateCheckItemReq) (*types.CreateCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	vo, err := CreateCheckItemReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	e, err := impl.taskDomain.CreateCheckItem(ctx, userId, vo)
	if err != nil {
		return nil, err
	}
	return CheckItemEntityToCreateRes(e), nil
}

func (impl *TaskAppImpl) UpdateCheckItem(ctx context.Context, itemId string, req *types.UpdateCheckItemReq) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	id64, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		return errors.New("检查事项 ID 格式错误")
	}
	vo, err := UpdateCheckItemReqToVO(req)
	if err != nil {
		return err
	}
	return impl.taskDomain.UpdateCheckItem(ctx, userId, id64, vo)
}

func (impl *TaskAppImpl) DeleteCheckItem(ctx context.Context, itemId string) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	id64, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		return errors.New("检查事项 ID 格式错误")
	}
	return impl.taskDomain.DeleteCheckItem(ctx, userId, id64)
}

func (impl *TaskAppImpl) ListCheckItems(ctx context.Context, taskId string) (types.ListCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 格式错误")
	}
	items, err := impl.taskDomain.ListCheckItems(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	return CheckItemEntitiesToReses(items), nil
}

func (impl *TaskAppImpl) BatchUpdateCheckItems(ctx context.Context, req *types.BatchUpdateCheckItemReq) (*types.BatchUpdateCheckItemRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	vos, err := BatchUpdateCheckItemReqToVOs(req)
	if err != nil {
		return nil, err
	}
	items, err := impl.taskDomain.BatchUpdateCheckItems(ctx, userId, vos)
	if err != nil {
		return nil, err
	}
	resList := CheckItemEntitiesToReses(items)
	return &types.BatchUpdateCheckItemRes{UpdatedCount: int64(len(resList)), Events: resList}, nil
}

// === Comment ===

func (impl *TaskAppImpl) GetCommentById(ctx context.Context, commentId string) (*types.CommentRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	id64, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return nil, errors.New("评论 ID 无效")
	}
	e, err := impl.taskDomain.GetCommentById(ctx, userId, id64)
	if err != nil {
		return nil, err
	}
	return CommentEntityToRes(e), nil
}

func (impl *TaskAppImpl) CreateComment(ctx context.Context, req *types.CreateCommentReq) (*types.CommentRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	vo, err := CreateCommentReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	e, err := impl.taskDomain.CreateComment(ctx, userId, vo)
	if err != nil {
		return nil, err
	}
	return CommentEntityToRes(e), nil
}

func (impl *TaskAppImpl) UpdateComment(ctx context.Context, commentId string, req *types.UpdateCommentReq) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	id64, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.New("评论 ID 无效")
	}
	vo, err := UpdateCommentReqToVO(req)
	if err != nil {
		return err
	}
	return impl.taskDomain.UpdateComment(ctx, userId, id64, vo)
}

func (impl *TaskAppImpl) DeleteComment(ctx context.Context, commentId string) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	id64, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.New("评论 ID 无效")
	}
	return impl.taskDomain.DeleteComment(ctx, userId, id64)
}

func (impl *TaskAppImpl) ListComments(ctx context.Context, taskId string) ([]*types.CommentRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 无效")
	}
	entities, err := impl.taskDomain.ListComments(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	return CommentEntitiesToListRes(entities), nil
}

func (impl *TaskAppImpl) SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error {
	return impl.taskDomain.SyncCommentUserProfile(ctx, userId, nickname, avatar)
}
