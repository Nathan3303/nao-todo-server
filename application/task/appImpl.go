package task

import (
	"context"
	"fmt"
	"naotodoserver/application/idutil"
	"naotodoserver/application/task/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/service"
	"naotodoserver/domain/task/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"time"
)

// NewTaskApp 创建任务应用层实例
func NewTaskApp(
	taskDomain service.TaskDomain,
	taskRepo repositories.Task,
	checkItemRepo repositories.TaskCheckItem,
	commentRepo repositories.TaskComment,
	publisher domaintypes.NotificationPublisher,
) *TaskAppImpl {
	impl := &TaskAppImpl{
		taskDomain:    taskDomain,
		taskRepo:      taskRepo,
		checkItemRepo: checkItemRepo,
		commentRepo:   commentRepo,
		publisher:     publisher,
	}
	return impl
}

// GetTaskById 获取单个任务信息
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务 ID
// @param includeDeleted 为 true 时可查询到已软删除的任务
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) GetTaskById(
	ctx context.Context,
	userId int64,
	taskId string,
	includeDeleted bool,
) (*dto.GetTaskRes, error) {
	// 1. 转换待办任务 ID
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	// 2. 调用仓库获取待办任务信息
	taskEntity, err := taskApp.taskRepo.GetById(ctx, userId, taskId64, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("GetTaskById: %w", err)
	}
	// 3. 转换为响应对象
	res := TaskEntityToGetRes(taskEntity)
	return res, nil
}

// CreateTask 创建任务
// @param ctx 上下文
// @param userId 用户 ID
// @param req 创建任务请求
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) CreateTask(
	ctx context.Context,
	userId int64,
	req *dto.CreateTaskReq,
) (*dto.GetTaskRes, error) {
	// 1. 请求体转换值对象
	createTaskValueObject, err := CreateTaskReqToValueObject(userId, req)
	if err != nil {
		return nil, err
	}
	taskEntity, err := taskApp.taskDomain.CreateTask(ctx, userId, createTaskValueObject)
	if err != nil {
		return nil, fmt.Errorf("CreateTask: %w", err)
	}
	// 3. 转换为响应对象
	res := TaskEntityToGetRes(taskEntity)
	return res, nil
}

// UpdateTask 更新任务
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务 ID
// @param req 更新任务请求
// @return 更新任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) UpdateTask(
	ctx context.Context,
	userId int64,
	taskId string,
	req *dto.UpdateTaskReq,
) error {
	// 1. 转换待办任务 ID
	taskIdInt64, err := idutil.ParseID(taskId)
	if err != nil {
		return domerr.ErrInvalidTaskID
	}
	// 2. 请求体转换值对象
	updateTaskValueObject, err := UpdateTaskReqToValueObject(userId, req)
	if err != nil {
		return err
	}
	// 3. 读-改-写：状态迁移与时间字段，由实体状态机/业务方法维护
	needReadWrite := req.State != nil ||
		req.ArchivedAt != nil ||
		req.StarMarkAt != nil ||
		req.GivenUpAt != nil
	var taskEntity *entities.Task
	if needReadWrite {
		taskEntity, err = taskApp.taskRepo.GetById(ctx, userId, taskIdInt64, false)
		if err != nil {
			return fmt.Errorf("UpdateTask: %w", err)
		}
		// 3.1 状态迁移：由实体状态机维护（含完成时间盖章/清空），此处仅作持久化翻译
		if req.State != nil {
			parsed, _ := entities.ParseTaskState(*req.State)
			if err := taskEntity.ChangeState(parsed); err != nil {
				return err
			}
			newState := taskEntity.State
			updateTaskValueObject.State = &newState
			updateTaskValueObject.CompletedAt = taskEntity.CompletedAt
		}
		// 3.2 归档时间
		if req.ArchivedAt != nil {
			if *req.ArchivedAt == "" {
				taskEntity.Unarchive()
				updateTaskValueObject.ArchivedAt = domaintypes.NewNullableTimeByTimeStrPtr(
					req.ArchivedAt,
				)
			} else {
				parsedTime, _ := time.Parse(time.RFC3339, *req.ArchivedAt)
				if parsedTime.IsZero() {
					parsedTime, _ = time.Parse("2006-01-02T15:04", *req.ArchivedAt)
				}
				taskEntity.Archive(&parsedTime)
				updateTaskValueObject.ArchivedAt = domaintypes.NewNullableTimeByTimeStrPtr(
					req.ArchivedAt,
				)
			}
		}
		// 3.3 收藏时间
		if req.StarMarkAt != nil {
			if *req.StarMarkAt == "" {
				// 空串：取消收藏（清空实体，置 null 落库，不再校验与开始时间的关系）
				taskEntity.Unstar()
				updateTaskValueObject.StarMarkAt = domaintypes.NewNullableTimeByTimeStrPtr(
					req.StarMarkAt,
				)
			} else {
				parsedTime, _ := time.Parse(time.RFC3339, *req.StarMarkAt)
				if parsedTime.IsZero() {
					parsedTime, _ = time.Parse("2006-01-02T15:04", *req.StarMarkAt)
				}
				taskEntity.ToggleStar(&parsedTime)
				updateTaskValueObject.StarMarkAt = domaintypes.NewNullableTimeByTimeStrPtr(
					req.StarMarkAt,
				)
			}
		}
		// 3.4 放弃时间
		if req.GivenUpAt != nil {
			if *req.GivenUpAt == "" {
				taskEntity.UngiveUp()
				updateTaskValueObject.GivenUpAt = domaintypes.NewNullableTimeByTimeStrPtr(
					req.GivenUpAt,
				)
			} else {
				parsedTime, _ := time.Parse(time.RFC3339, *req.GivenUpAt)
				if parsedTime.IsZero() {
					parsedTime, _ = time.Parse("2006-01-02T15:04", *req.GivenUpAt)
				}
				taskEntity.GiveUp(&parsedTime)
				updateTaskValueObject.GivenUpAt = domaintypes.NewNullableTimeByTimeStrPtr(
					req.GivenUpAt,
				)
			}
		}
		// 3.5 实体时间参数校验
		if err := taskEntity.IsDatesValid(); err != nil {
			return err
		}
	}
	if err := taskApp.taskRepo.Update(ctx, userId, taskIdInt64, updateTaskValueObject); err != nil {
		return fmt.Errorf("UpdateTask: %w", err)
	}
	return nil
}

// DeleteTask 删除任务
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务 ID
// @return 删除任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) DeleteTask(
	ctx context.Context,
	userId int64,
	taskId string,
) error {
	// 1. 转换待办任务 ID
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
// @param userId 用户 ID
// @param taskId 任务 ID
// @return 恢复任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) RestoreTask(
	ctx context.Context,
	userId int64,
	taskId string,
) error {
	// 1. 转换待办任务 ID
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
// @param userId 用户 ID
// @param taskId 任务 ID
// @return 任务响应
// @return error 错误信息
func (taskApp *TaskAppImpl) CopyTask(
	ctx context.Context,
	userId int64,
	taskId string,
) (*dto.GetTaskRes, error) {
	// 1. 转换待办任务 ID
	taskId64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	// 2. 调用领域层复制任务
	taskEntity, err := taskApp.taskDomain.Copy(ctx, userId, taskId64)
	if err != nil {
		return nil, fmt.Errorf("CopyTask: %w", err)
	}
	// 3. 转换为响应对象
	res := TaskEntityToGetRes(taskEntity)
	return res, nil
}

// ListTask 获取任务列表
// @param ctx 上下文
// @param userId 用户 ID
// @param req 获取任务列表请求
// @return 任务列表响应
// @return error 错误信息
func (taskApp *TaskAppImpl) ListTask(
	ctx context.Context,
	userId int64,
	req *dto.ListTaskReq,
) (dto.ListTaskRes, *dto.Pagination, error) {
	// 1. 请求体转换值对象
	queryTaskValueObject, err := ListTaskReqToQueryTaskValueObject(userId, req)
	if err != nil {
		return nil, nil, fmt.Errorf("ListTask req: %w", err)
	}
	// 2. 分页值对象
	paginationValueObject := valueobjects.NewPagination(0, req.Page, req.Limit)
	// 3. 调用领域层获取任务列表
	taskEntities, paginationValueObject, err := taskApp.taskDomain.List(
		ctx,
		userId,
		queryTaskValueObject,
		paginationValueObject,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("ListTask: %w", err)
	}
	// 4. 转换为响应对象
	tasks := TaskEntitiesToGetReses(taskEntities)
	paginationRes := PaginationValueObjectToRes(paginationValueObject)
	return tasks, paginationRes, nil
}

// ListTaskSync 增量同步任务列表
// 包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
func (taskApp *TaskAppImpl) ListTaskSync(
	ctx context.Context,
	userId int64,
	req *dto.ListTaskReq,
) (dto.ListTaskRes, error) {
	cursor, err := idutil.ParseUpdatedAtCursor(req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	var cursorID int64
	if req.CursorId != "" {
		cursorID, err = idutil.ParseID(req.CursorId)
		if err != nil {
			return nil, fmt.Errorf("cursorId 格式错误: %w", err)
		}
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	taskEntities, err := taskApp.taskDomain.ListSync(ctx, userId, cursor, cursorID, limit)
	if err != nil {
		return nil, fmt.Errorf("ListTaskSync: %w", err)
	}
	return TaskEntitiesToGetReses(taskEntities), nil
}

// SnoozeTask 稍后提醒
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务ID
// @param req 稍后提醒请求
// @return 稍后提醒响应
// @return error 错误
func (taskApp *TaskAppImpl) SnoozeTask(
	ctx context.Context,
	userId int64,
	taskId string,
	req *dto.SnoozeTaskReq,
) (*dto.SnoozeTaskRes, error) {
	// 1. 转换任务 ID
	taskIdInt64, err := idutil.ParseID(taskId)
	if err != nil {
		return nil, domerr.ErrInvalidTaskID
	}
	// 2. 调用领域层设置稍后提醒
	newRemindAt, err := taskApp.taskDomain.Snooze(
		ctx,
		userId,
		taskIdInt64,
		req.DurationMinutes,
	)
	if err != nil {
		return nil, fmt.Errorf("SnoozeTask: %w", err)
	}
	// 3. 返回结果
	return &dto.SnoozeTaskRes{
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
	for _, task := range tasks {
		if err := taskApp.publisher.PublishReminder(
			ctx,
			int64(task.UserId),
			domaintypes.ReminderEvent{
				Type:        "REMINDER",
				TaskId:      idutil.FormatID(task.Id),
				TaskName:    task.Name,
				Description: task.Description,
				RemindAt:    task.RemindAt.ToString(time.RFC3339),
			},
		); err != nil {
			return fmt.Errorf("ProcessReminders publish: %w", err)
		}
	}
	return nil
}

// --- TaskCheckItem ---

// GetTaskCheckItemById 获取检查事项详情
// @param ctx 上下文
// @param userId 用户 ID
// @param itemId 检查事项 ID
// @return 检查事项响应
// @return error 错误信息
func (impl *TaskAppImpl) GetTaskCheckItemById(
	ctx context.Context,
	userId int64,
	itemId string,
) (*dto.GetTaskCheckItemRes, error) {
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
// @param userId 用户 ID
// @param req 创建检查事项请求
// @return 创建检查事项响应
// @return error 错误信息
func (impl *TaskAppImpl) CreateTaskCheckItem(
	ctx context.Context,
	userId int64,
	req *dto.CreateTaskCheckItemReq,
) (*dto.CreateTaskCheckItemRes, error) {
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
// @param userId 用户 ID
// @param itemId 检查事项 ID
// @param req 更新检查事项请求
// @return error 错误信息
func (impl *TaskAppImpl) UpdateTaskCheckItem(
	ctx context.Context,
	userId int64,
	itemId string,
	req *dto.UpdateTaskCheckItemReq,
) error {
	id64, err := idutil.ParseID(itemId)
	if err != nil {
		return domerr.ErrInvalidItemID
	}
	vo, err := UpdateTaskCheckItemReqToVO(req)
	if err != nil {
		return err
	}
	// 状态迁移：读-改-写，复用实体 SetDone 保持状态决策在领域层
	if vo.IsDone != nil {
		item, err := impl.checkItemRepo.GetCheckItemById(ctx, userId, id64)
		if err != nil {
			return fmt.Errorf("UpdateCheckItem: %w", err)
		}
		item.SetDone(*vo.IsDone)
		newIsDone := item.IsDone
		vo.IsDone = &newIsDone
	}
	if err := impl.checkItemRepo.UpdateCheckItem(ctx, userId, id64, vo); err != nil {
		return fmt.Errorf("UpdateCheckItem: %w", err)
	}
	return nil
}

// DeleteTaskCheckItem 删除检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param itemId 检查事项 ID
// @return error 错误信息
func (impl *TaskAppImpl) DeleteTaskCheckItem(
	ctx context.Context,
	userId int64,
	itemId string,
) error {
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
// @param userId 用户 ID
// @param taskId 待办任务 ID
// @return 检查事项列表响应
// @return error 错误信息
func (impl *TaskAppImpl) ListTaskCheckItems(
	ctx context.Context,
	userId int64,
	taskId string,
) (dto.ListTaskCheckItemRes, error) {
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

// ListTaskCheckItemSync 增量同步检查事项列表
// 包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
func (impl *TaskAppImpl) ListTaskCheckItemSync(
	ctx context.Context,
	userId int64,
	updatedAt string,
	cursorId string,
	limit int,
) (dto.ListTaskCheckItemRes, error) {
	cursor, err := idutil.ParseUpdatedAtCursor(updatedAt)
	if err != nil {
		return nil, err
	}
	var cursorID int64
	if cursorId != "" {
		cursorID, err = idutil.ParseID(cursorId)
		if err != nil {
			return nil, fmt.Errorf("cursorId 格式错误: %w", err)
		}
	}
	if limit <= 0 {
		limit = 100
	}
	items, err := impl.checkItemRepo.ListCheckItemsSync(ctx, userId, cursor, cursorID, limit)
	if err != nil {
		return nil, fmt.Errorf("ListTaskCheckItemSync: %w", err)
	}
	return TaskCheckItemEntitiesToReses(items), nil
}

// BatchUpdateTaskCheckItems 批量更新检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param req 批量更新检查事项请求
// @return 批量更新检查事项响应
// @return error 错误信息
func (impl *TaskAppImpl) BatchUpdateTaskCheckItems(
	ctx context.Context,
	userId int64,
	req *dto.BatchUpdateTaskCheckItemReq,
) (*dto.BatchUpdateTaskCheckItemRes, error) {
	vos, err := BatchUpdateTaskCheckItemReqToVOs(req)
	if err != nil {
		return nil, err
	}
	// 状态迁移：任一检查项携带 IsDone 时，先读-改-写复用实体方法
	needReadWrite := false
	for _, vo := range vos {
		if vo.IsDone != nil {
			needReadWrite = true
			break
		}
	}
	if needReadWrite {
		for _, vo := range vos {
			if vo.IsDone == nil {
				continue
			}
			item, err := impl.checkItemRepo.GetCheckItemById(ctx, userId, vo.Id)
			if err != nil {
				return nil, fmt.Errorf("BatchUpdateCheckItems: %w", err)
			}
			item.SetDone(*vo.IsDone)
			newIsDone := item.IsDone
			vo.IsDone = &newIsDone
		}
	}
	items, err := impl.checkItemRepo.BatchUpdateCheckItems(ctx, userId, vos)
	if err != nil {
		return nil, fmt.Errorf("BatchUpdateCheckItems: %w", err)
	}
	resList := TaskCheckItemEntitiesToReses(items)
	return &dto.BatchUpdateTaskCheckItemRes{
		UpdatedCount: int64(len(resList)),
		Events:       resList,
	}, nil
}

// --- Comment ---

// GetTaskCommentById 获取评论详情
// @param ctx 上下文
// @param userId 用户 ID
// @param commentId 评论 ID
// @return 评论响应
// @return error 错误信息
func (impl *TaskAppImpl) GetTaskCommentById(
	ctx context.Context,
	userId int64,
	commentId string,
) (*dto.TaskCommentRes, error) {
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
// @param userId 用户 ID
// @param req 创建评论请求
// @return 创建评论响应
// @return error 错误信息
func (impl *TaskAppImpl) CreateTaskComment(
	ctx context.Context,
	userId int64,
	req *dto.CreateTaskCommentReq,
) (*dto.TaskCommentRes, error) {
	vo, err := CreateTaskCommentReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	// 幂等创建：客户端指定 id 时走 upsert（LWW + create 冲突检测）
	e, _, err := impl.commentRepo.UpsertComment(ctx, userId, vo)
	if err != nil {
		return nil, fmt.Errorf("CreateComment: %w", err)
	}
	return TaskCommentEntityToRes(e), nil
}

// UpdateTaskComment 更新评论
// @param ctx 上下文
// @param userId 用户 ID
// @param commentId 评论 ID
// @param req 更新评论请求
// @return error 错误信息
func (impl *TaskAppImpl) UpdateTaskComment(
	ctx context.Context,
	userId int64,
	commentId string,
	req *dto.UpdateTaskCommentReq,
) error {
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
// @param userId 用户 ID
// @param commentId 评论 ID
// @return error 错误信息
func (impl *TaskAppImpl) DeleteTaskComment(
	ctx context.Context,
	userId int64,
	commentId string,
) error {
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
// @param userId 用户 ID
// @param taskId 待办任务 ID
// @return 评论列表响应
// @return error 错误信息
func (impl *TaskAppImpl) ListTaskComments(
	ctx context.Context,
	userId int64,
	taskId string,
) ([]*dto.TaskCommentRes, error) {
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

// ListTaskCommentSync 增量同步评论列表
// 包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
func (impl *TaskAppImpl) ListTaskCommentSync(
	ctx context.Context,
	userId int64,
	updatedAt string,
	cursorId string,
	limit int,
) ([]*dto.TaskCommentRes, error) {
	cursor, err := idutil.ParseUpdatedAtCursor(updatedAt)
	if err != nil {
		return nil, err
	}
	var cursorID int64
	if cursorId != "" {
		cursorID, err = idutil.ParseID(cursorId)
		if err != nil {
			return nil, fmt.Errorf("cursorId 格式错误: %w", err)
		}
	}
	if limit <= 0 {
		limit = 100
	}
	commentEntities, err := impl.commentRepo.ListCommentsSync(ctx, userId, cursor, cursorID, limit)
	if err != nil {
		return nil, fmt.Errorf("ListTaskCommentSync: %w", err)
	}
	return TaskCommentEntitiesToListRes(commentEntities), nil
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
