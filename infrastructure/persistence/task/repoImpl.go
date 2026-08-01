package task

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	query "naotodoserver/infrastructure/utils/query"
	"time"

	"gorm.io/gorm"
)

type TaskRepoImpl struct {
	db *gorm.DB
}

// 编译期接口实现断言
var _ repositories.Task = (*TaskRepoImpl)(nil)
var _ repositories.TaskCheckItem = (*TaskRepoImpl)(nil)
var _ repositories.TaskComment = (*TaskRepoImpl)(nil)

func NewTaskRepo(db *gorm.DB) *TaskRepoImpl {
	return &TaskRepoImpl{db: db}
}

// GetById 根据任务ID获取任务信息
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @param includeDeleted 为 true 时可查询到已软删除的任务
// @return 任务实体
// @return error 错误
func (taskRepo *TaskRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	taskId int64,
	includeDeleted bool,
) (*entities.Task, error) {
	// 1. 创建结果模型
	taskModel := &models.Task{}
	// 2. 转换查询实体到模型
	var whereCond models.Task
	whereCond.UserId = userId
	whereCond.ID = taskId
	// 3. 查询
	db := taskRepo.db.WithContext(ctx)
	if includeDeleted {
		db = db.Unscoped()
	}
	tx := db.Where(whereCond).First(taskModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 转换模型到实体并返回
	return TaskModel2Entity(taskModel), nil
}

// Create 创建任务
// @param ctx 上下文
// @param createEntity 创建实体
// @return 任务实体
// @return error 错误
func (taskRepo *TaskRepoImpl) Create(
	ctx context.Context,
	userId int64,
	createTaskValueObject *valueobjects.CreateTask,
) (*entities.Task, error) {
	// 1. 转换创建实体到模型
	createModel := CreateTaskValueObjectToModel(userId, createTaskValueObject)
	// 2. 创建
	tx := taskRepo.db.WithContext(ctx).Create(createModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 转换模型到实体并返回
	return TaskModel2Entity(createModel), nil
}

// GetMaxSortId 获取任务最大排序 ID
// @param ctx 上下文
// @param userId 用户ID
// @return 最大排序ID
func (taskRepo *TaskRepoImpl) GetMaxSortId(ctx context.Context, userId int64) uint16 {
	var maxSortId uint16 = 255
	taskRepo.db.
		WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ?", userId).
		Pluck("MAX(sort_id)", &maxSortId)
	return maxSortId
}

// Update 更新任务
// @param ctx 上下文
// @param whereEntity 查询实体
// @param updateEntity 更新实体
// @return error 错误
func (taskRepo *TaskRepoImpl) Update(
	ctx context.Context,
	userId int64,
	taskId int64,
	updateTaskValueObject *valueobjects.UpdateTask,
) error {
	// 1. 转换查询实体到模型
	var whereCond models.Task
	whereCond.UserId = userId
	whereCond.ID = taskId
	// 2. 转换更新实体到 map
	updateMap := UpdateTaskValueObjectToMap(updateTaskValueObject)
	// 3. 更新
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Updates(updateMap)
	return tx.Error
}

// Delete 删除任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskRepo *TaskRepoImpl) Delete(ctx context.Context, userId int64, taskId int64) error {
	// 1. 转换查询实体到模型
	var whereCond models.Task
	whereCond.UserId = userId
	whereCond.ID = taskId
	// 2. 删除
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Delete(&models.Task{})
	return tx.Error
}

// Restore 恢复任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskRepo *TaskRepoImpl) Restore(ctx context.Context, userId int64, taskId int64) error {
	// 1. 转换查询实体到模型
	var whereCond models.Task
	whereCond.UserId = userId
	whereCond.ID = taskId
	// 2. 恢复
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).Unscoped().
		Where(whereCond).
		Update("deleted_at", nil)
	return tx.Error
}

// GetWeekRange 获取某时间所在周的开始（周一）和结束（周日）
// @param t 时间值
// @return start 周一时间
// @return end 周日时间
func GetWeekRange(t time.Time) (start, end time.Time) {
	// 将时间调整到 UTC 或本地时区（根据你的数据库时区设置）
	loc := time.Local // 或 time.UTC
	t = t.In(loc)
	// 计算距离周一的天数（Go 中 Weekday() 返回 0=Sunday, 1=Monday, ..., 6=Saturday）
	offset := int(t.Weekday())
	if offset == 0 {
		offset = 7 // 如果是周日，则往前推 6 天到周一
	}
	start = t.AddDate(0, 0, -offset+1) // 周一
	end = start.AddDate(0, 0, 6)       // 周日
	// 设置时间为当天的 00:00:00 到 23:59:59
	start = time.Date(
		start.Year(),
		start.Month(),
		start.Day(), 0, 0, 0, 0, start.Location())
	end = time.Date(end.Year(),
		end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
	return
}

// List 获取任务列表
// @param ctx 上下文
// @param userId 用户ID
// @param q 查询值对象
// @param pagination 分页值对象
// @return 任务实体列表
// @return error 错误
func (taskRepo *TaskRepoImpl) List(
	ctx context.Context,
	userId int64,
	q *valueobjects.QueryTask,
	pagination *valueobjects.Pagination,
) ([]*entities.Task, *valueobjects.Pagination, error) {
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where("user_id = ?", userId).
		Scopes(
			ByParentTaskId(q.ParentTaskId),
			ByProjectOrTag(q),
			ByTaskName(q.Name),
			ByTaskDescription(q.Description),
			ByTaskState(q.State),
			ByTaskPriority(q.Priority),
			ByTaskTimeRange(q.StartAt, q.EndAt),
			ByTaskDeleted(q.IsDeleted),
			ByTaskArchived(q.IsArchived),
			ByTaskStarMarked(q.IsStarMarked),
			ByTaskGivenUpFlag(q.IsGivenUp),
			ByRelativeDate(q.RelativeDate),
			query.Sort(q.Sort),
		)

	var total int64
	tx.Count(&total)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}

	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}

	var taskModels []*models.Task
	tx = tx.
		Scopes(query.Paginate(pagination.Page, pagination.Limit)).
		Find(&taskModels)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	pagination.Total = total
	return TaskModels2Entities(taskModels), pagination, nil
}

// --- 任务提醒相关 ---

// Snooze 稍后提醒
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @param remindAt 新提醒时间
// @return error 错误
func (taskRepo *TaskRepoImpl) Snooze(
	ctx context.Context,
	userId int64,
	taskId int64,
	remindAt string,
) error {
	// 1. 构建查询条件
	var whereCond models.Task
	whereCond.UserId = userId
	whereCond.ID = taskId
	// 2. 更新 remind_at
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Update("remind_at", remindAt)
	return tx.Error
}

// GetDueReminders 获取到期提醒任务
// @param ctx 上下文
// @return 任务实体列表
// @return error 错误
func (taskRepo *TaskRepoImpl) GetDueReminders(ctx context.Context) ([]*entities.Task, error) {
	// 1. 确保上下文非空
	if ctx == nil {
		ctx = context.Background()
	}
	// 2. 查询到期提醒
	var taskModels []*models.Task
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where("remind_at <= NOW()").
		Where("remind_at IS NOT NULL").
		Find(&taskModels)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 转换模型到实体
	return TaskModels2Entities(taskModels), nil
}

// ClearRemindRepeat 清除提醒重复规则
// @param ctx 上下文
// @param taskId 任务ID
// @return error 错误
func (taskRepo *TaskRepoImpl) ClearRemindRepeat(ctx context.Context, taskId int64) error {
	// 1. 确保上下文非空
	if ctx == nil {
		ctx = context.Background()
	}
	// 2. 构建查询条件
	var whereCond models.Task
	whereCond.ID = taskId
	// 3. 更新
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Updates(map[string]interface{}{
			"remind_repeat":   0,
			"remind_at":       nil,
			"remind_time":     "",
			"remind_weekdays": 0,
		})
	return tx.Error
}

// UpdateRemindAt 更新提醒时间
// @param ctx 上下文
// @param taskId 任务ID
// @param remindAt 新提醒时间，空字符串表示清除
// @return error 错误
func (taskRepo *TaskRepoImpl) UpdateRemindAt(
	ctx context.Context,
	taskId int64,
	remindAt string,
) error {
	// 1. 确保上下文非空
	if ctx == nil {
		ctx = context.Background()
	}
	// 2. 构建查询条件
	var whereCond models.Task
	whereCond.ID = taskId
	// 3. 更新
	if remindAt == "" {
		tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
			Where(whereCond).
			Update("remind_at", nil)
		return tx.Error
	}
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Update("remind_at", remindAt)
	return tx.Error
}

// --- 任务检查项相关 ---

// GetCheckItemById 获取任务检查项
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 检查项ID
// @return 任务检查项实体
// @return error 错误
func (repo *TaskRepoImpl) GetCheckItemById(
	ctx context.Context,
	userId int64,
	checkItemId int64,
) (*entities.TaskCheckItem, error) {
	var m models.TaskCheckItem
	tx := repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("id = ? AND user_id = ?", checkItemId, userId).First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCheckItemModel2Entity(&m), nil
}

// CreateCheckItem 创建任务检查项
// @param ctx 上下文
// @param userId 用户ID
// @param vo 创建任务检查项值对象
// @return 任务检查项实体
// @return error 错误
func (repo *TaskRepoImpl) CreateCheckItem(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskCheckItem,
) (*entities.TaskCheckItem, error) {
	m := TaskCheckItemValueObjectToModel(vo)
	tx := repo.db.WithContext(ctx).Model(&models.TaskCheckItem{}).Create(m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCheckItemModel2Entity(m), nil
}

// UpdateCheckItem 更新任务检查项
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 检查项ID
// @param vo 更新任务检查项值对象
// @return error 错误
func (repo *TaskRepoImpl) UpdateCheckItem(
	ctx context.Context,
	userId int64,
	checkItemId int64,
	vo *valueobjects.UpdateTaskCheckItem,
) error {
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("id = ? AND user_id = ?", checkItemId, userId).
		Updates(UpdateTaskCheckItemValueObjectToMap(vo)).
		Error
}

// DeleteCheckItem 删除任务检查项
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 检查项ID
// @return error 错误
func (repo *TaskRepoImpl) DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error {
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("id = ? AND user_id = ?", checkItemId, userId).
		Delete(&models.TaskCheckItem{}).
		Error
}

// ListCheckItems 获取任务检查项列表
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 待办任务任务ID
// @return 任务检查项实体列表
// @return error 错误
func (repo *TaskRepoImpl) ListCheckItems(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.TaskCheckItem, error) {
	var list []*models.TaskCheckItem
	tx := repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("user_id = ? AND task_id = ?", userId, taskId).
		Find(&list)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCheckItemModels2Entities(list), nil
}

// GetMaxCheckItemSortId 获取任务检查项最大排序ID
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 待办任务任务ID
// @return 最大排序ID
// @return error 错误
func (repo *TaskRepoImpl) GetMaxCheckItemSortId(ctx context.Context, userId, taskId int64) uint16 {
	var maxSortId uint16 = 255
	repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("user_id = ? AND task_id = ?", userId, taskId).
		Pluck("MAX(sort_id)", &maxSortId)
	return maxSortId
}

// BatchUpdateCheckItems 批量更新任务检查项
// @param ctx 上下文
// @param userId 用户ID
// @param vos 批量更新任务检查项值对象列表
// @return 任务检查项实体列表
// @return error 错误
func (repo *TaskRepoImpl) BatchUpdateCheckItems(
	ctx context.Context,
	userId int64,
	vos []*valueobjects.BatchUpdateTaskCheckItem,
) ([]*entities.TaskCheckItem, error) {
	tx := repo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	IDs := make([]int64, 0, len(vos))
	var err error
	for _, vo := range vos {
		err = tx.
			Model(&models.TaskCheckItem{}).
			Where("id = ? AND user_id = ?", vo.Id, userId).
			Updates(BatchUpdateTaskCheckItemValueObjectToMap(vo)).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		IDs = append(IDs, vo.Id)
	}
	var updated []*models.TaskCheckItem
	err = tx.
		Model(&models.TaskCheckItem{}).
		Where("id IN ? AND user_id = ?", IDs, userId).
		Find(&updated).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return TaskCheckItemModels2Entities(updated), nil
}

// --- 任务评论相关 ---

// GetCommentById 获取任务评论
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @return 任务实体
// @return error 错误
func (repo *TaskRepoImpl) GetCommentById(
	ctx context.Context,
	userId int64,
	commentId int64,
) (*entities.TaskComment, error) {
	var m models.TaskComment
	tx := repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ? AND id = ?", userId, commentId).
		First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCommentModel2Entity(&m), nil
}

// CreateComment 创建任务评论
// @param ctx 上下文
// @param userId 用户ID
// @param vo 创建任务评论值对象
// @return 任务实体
// @return error 错误
func (repo *TaskRepoImpl) CreateComment(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskComment,
) (*entities.TaskComment, error) {
	var user models.User
	var err error
	err = repo.db.
		WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userId).
		First(&user).
		Error
	if err != nil {
		return nil, err
	}
	m := CreateTaskCommentValueObjectToModel(vo)
	m.Nickname = user.Nickname
	m.Avatar = user.Avatar
	err = repo.db.WithContext(ctx).Create(m).Error
	if err != nil {
		return nil, err
	}
	return TaskCommentModel2Entity(m), nil
}

// UpdateComment 更新任务评论
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @param vo 更新任务评论值对象
// @return error 错误
func (repo *TaskRepoImpl) UpdateComment(
	ctx context.Context,
	userId int64,
	commentId int64,
	vo *valueobjects.UpdateTaskComment,
) error {
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ? AND id = ?", userId, commentId).
		Updates(UpdateTaskCommentValueObjectToMap(vo)).
		Error
}

// DeleteComment 删除任务评论
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @return error 错误
func (repo *TaskRepoImpl) DeleteComment(ctx context.Context, userId, commentId int64) error {
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ? AND id = ?", userId, commentId).
		Delete(&models.TaskComment{}).
		Error
}

// ListComments 获取任务评论列表
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 待办任务任务ID
// @return 任务评论实体列表
// @return error 错误
func (repo *TaskRepoImpl) ListComments(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.TaskComment, error) {
	var list []*models.TaskComment
	tx := repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ? AND task_id = ?", userId, taskId).
		Find(&list)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCommentModels2Entities(list), nil
}

// SyncCommentUserProfile 同步任务评论用户属性
// @param ctx 上下文
// @param userId 用户ID
// @param nickname 昵称
// @param avatar 头像
// @return error 错误
func (repo *TaskRepoImpl) SyncCommentUserProfile(
	ctx context.Context,
	userId int64,
	nickname, avatar string,
) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return nil
	}
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ?", userId).
		Updates(updates).
		Error
}

// SoftDeleteByProjectId 软删除指定项目下的所有任务
// 用于项目删除时的级联操作，设置 deleted_at 为当前时间
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) SoftDeleteByProjectId(ctx context.Context, userId int64, projectId int64) error {
	return dbs.DBFrom(ctx, repo.db).WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Update("deleted_at", time.Now()).Error
}

// RestoreByProjectId 恢复指定项目下的所有任务
// 用于项目恢复时的级联操作，清除 deleted_at
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) RestoreByProjectId(ctx context.Context, userId int64, projectId int64) error {
	return dbs.DBFrom(ctx, repo.db).WithContext(ctx).
		Unscoped().
		Model(&models.Task{}).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Update("deleted_at", nil).Error
}

// ArchiveByProjectId 归档指定项目下的所有任务
// 用于项目归档时的级联操作，设置 archived_at 为当前时间
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) ArchiveByProjectId(ctx context.Context, userId int64, projectId int64) error {
	return dbs.DBFrom(ctx, repo.db).WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Update("archived_at", time.Now()).Error
}

// UnarchiveByProjectId 取消归档指定项目下的所有任务
// 用于项目取消归档时的级联操作，清除 archived_at
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) UnarchiveByProjectId(ctx context.Context, userId int64, projectId int64) error {
	return dbs.DBFrom(ctx, repo.db).WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Update("archived_at", nil).Error
}
