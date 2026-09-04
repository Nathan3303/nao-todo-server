package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	query "naotodoserver/infrastructure/utils/query"
	"strconv"
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

// Upsert 幂等写入任务：客户端指定 id 时创建或覆盖
// - 记录不存在：带 id 创建
// - 记录存在：create 语义冲突检测（请求携带 createdAt 且与库中 created_at 相差 > 1 分钟 → ErrIDConflict）；
//   LWW 判定：请求 updatedAt 更旧则 no-op 返回当前版本，否则覆盖（仅更新 Create VO 表达的字段 + updated_at）
func (taskRepo *TaskRepoImpl) Upsert(
	ctx context.Context,
	userId int64,
	createTaskValueObject *valueobjects.CreateTask,
) (*entities.Task, bool, error) {
	if createTaskValueObject.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now（createdAt 保留客户端值供冲突检测）
		createTaskValueObject.UpdatedAt = time.Now()
		entity, err := taskRepo.Create(ctx, userId, createTaskValueObject)
		return entity, true, err
	}
	var existing models.Task
	err := taskRepo.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", createTaskValueObject.Id, userId).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		createTaskValueObject.UpdatedAt = time.Now()
		entity, createErr := taskRepo.Create(ctx, userId, createTaskValueObject)
		return entity, true, createErr
	}
	if err != nil {
		return nil, false, err
	}
	existingEntity := TaskModel2Entity(&existing)
	outcome, err := types.DecideUpsert(
		existing.CreatedAt, existingEntity.UpdatedAt,
		createTaskValueObject.CreatedAt, createTaskValueObject.UpdatedAt,
		time.Minute,
	)
	if err != nil {
		return nil, false, err
	}
	if outcome == types.UpsertNoop {
		return existingEntity, false, nil
	}
	updateMap := CreateTaskVOToUpdateMap(createTaskValueObject)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now（LWW 判定仍用客户端时间）
	updateMap["updated_at"] = time.Now()
	// 覆盖已软删记录（墓碑）时复活：未携带删除时间时显式清 deleted_at，否则增量拉取仍视为删除
	if _, ok := updateMap["deleted_at"]; !ok {
		updateMap["deleted_at"] = gorm.Expr("NULL")
	}
	if err := taskRepo.db.WithContext(ctx).Unscoped().
		Model(&models.Task{}).
		Where("id = ? AND user_id = ?", createTaskValueObject.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	entity, err := taskRepo.GetById(ctx, userId, createTaskValueObject.Id, true)
	return entity, false, err
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
	// 3. LWW 乐观锁：请求 updatedAt 早于库中版本时不更新（防旧数据回滚）
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond)
	if !updateTaskValueObject.UpdatedAt.IsZero() {
		tx = tx.Where("updated_at <= ?", updateTaskValueObject.UpdatedAt)
	}
	// 4. 更新（map 不含 UpdatedAt，gorm 自动刷新为 now）
	tx = tx.Updates(updateMap)
	return tx.Error
}

// Delete 删除任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskRepo *TaskRepoImpl) Delete(ctx context.Context, userId int64, taskId int64) error {
	// 1. 软删同时推进 updated_at，保证删除墓碑可被增量拉取发现
	tx := taskRepo.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id = ? AND user_id = ?", taskId, userId).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
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
			ByProjects(q.ProjectIds),
			ByTags(q.TagIds),
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

// ListSync 增量同步列表：包含软删墓碑（Unscoped），(updated_at, id) keyset 游标 + 稳定排序 + limit
func (taskRepo *TaskRepoImpl) ListSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.Task, error) {
	scopes := []func(db *gorm.DB) *gorm.DB{
		query.ByKeysetCursor(cursor, cursorID),
		query.SyncOrder(),
	}
	tx := taskRepo.db.WithContext(ctx).Unscoped().
		Model(&models.Task{}).
		Where("user_id = ?", userId).
		Scopes(scopes...)

	if limit <= 0 {
		limit = 100
	}

	var taskModels []*models.Task
	tx = tx.Limit(limit).Find(&taskModels)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskModels2Entities(taskModels), nil
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
	// 2. 解析为 time.Time 绑定：RFC3339 字符串（如 UTC 下的 "...Z"）会被 MySQL 拒绝，
	//    由 driver 按 DATETIME 格式序列化则无此问题
	t, err := time.Parse(time.RFC3339, remindAt)
	if err != nil {
		return err
	}
	// 3. 更新 remind_at
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Update("remind_at", t)
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
// @param expectedRemindAt 期望的当前提醒时间（CAS：仅当 remind_at 仍为该值时清空，防与 Snooze 竞态）
// @return 是否实际变更（false 表示提醒已被改期/删除，调用方应跳过）
// @return error 错误
func (taskRepo *TaskRepoImpl) ClearRemindRepeat(
	ctx context.Context,
	taskId int64,
	expectedRemindAt time.Time,
) (bool, error) {
	// 1. 确保上下文非空
	if ctx == nil {
		ctx = context.Background()
	}
	// 2. CAS 更新：remind_at 仍是扫描时的值才清空（毫秒截断与 DATETIME(3) 精度对齐）
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where("id = ? AND remind_at = ?", taskId, expectedRemindAt.Truncate(time.Millisecond)).
		Updates(map[string]interface{}{
			"remind_repeat":   0,
			"remind_at":       nil,
			"remind_time":     "",
			"remind_weekdays": 0,
		})
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

// UpdateRemindAt 更新提醒时间
// @param ctx 上下文
// @param taskId 任务ID
// @param expectedRemindAt 期望的当前提醒时间（CAS：仅当 remind_at 仍为该值时更新，防与 Snooze 竞态）
// @param remindAt 新提醒时间，空字符串表示清除
// @return 是否实际变更（false 表示提醒已被改期/删除，调用方应跳过）
// @return error 错误
func (taskRepo *TaskRepoImpl) UpdateRemindAt(
	ctx context.Context,
	taskId int64,
	expectedRemindAt time.Time,
	remindAt string,
) (bool, error) {
	// 1. 确保上下文非空
	if ctx == nil {
		ctx = context.Background()
	}
	// 2. CAS 更新：remind_at 仍是扫描时的值才更新
	//    RFC3339 字符串（如 UTC 下的 "...Z"）会被 MySQL 拒绝，解析为 time.Time 由 driver 序列化
	var value any
	if remindAt == "" {
		value = nil
	} else {
		t, err := time.Parse(time.RFC3339, remindAt)
		if err != nil {
			return false, err
		}
		value = t
	}
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where("id = ? AND remind_at = ?", taskId, expectedRemindAt.Truncate(time.Millisecond)).
		Update("remind_at", value)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
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

// UpsertCheckItem 幂等写入任务检查项：客户端指定 id 时创建或覆盖
// 语义与 Task.Upsert 一致（LWW + create 冲突检测）
func (repo *TaskRepoImpl) UpsertCheckItem(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskCheckItem,
) (*entities.TaskCheckItem, bool, error) {
	if vo.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, err := repo.CreateCheckItem(ctx, userId, vo)
		return entity, true, err
	}
	var existing models.TaskCheckItem
	err := repo.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, createErr := repo.CreateCheckItem(ctx, userId, vo)
		return entity, true, createErr
	}
	if err != nil {
		return nil, false, err
	}
	existingEntity := TaskCheckItemModel2Entity(&existing)
	outcome, err := types.DecideUpsert(
		existing.CreatedAt, existingEntity.UpdatedAt,
		vo.CreatedAt, vo.UpdatedAt,
		time.Minute,
	)
	if err != nil {
		return nil, false, err
	}
	if outcome == types.UpsertNoop {
		return existingEntity, false, nil
	}
	updateMap := TaskCheckItemVOToUpdateMap(vo)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now
	updateMap["updated_at"] = time.Now()
	// 覆盖已软删记录（墓碑）时复活：显式清 deleted_at
	updateMap["deleted_at"] = gorm.Expr("NULL")
	if err := repo.db.WithContext(ctx).Unscoped().
		Model(&models.TaskCheckItem{}).
		Where("id = ? AND user_id = ?", vo.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	var updated models.TaskCheckItem
	if err := repo.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&updated).Error; err != nil {
		return nil, false, err
	}
	return TaskCheckItemModel2Entity(&updated), false, nil
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
	tx := repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("id = ? AND user_id = ?", checkItemId, userId)
	// LWW 乐观锁：请求 updatedAt 早于库中版本时不更新
	if !vo.UpdatedAt.IsZero() {
		tx = tx.Where("updated_at <= ?", vo.UpdatedAt)
	}
	return tx.Updates(UpdateTaskCheckItemValueObjectToMap(vo)).Error
}

// DeleteCheckItem 删除任务检查项
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 检查项ID
// @return error 错误
func (repo *TaskRepoImpl) DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error {
	// 软删同时推进 updated_at，保证删除墓碑可被增量拉取发现
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskCheckItem{}).
		Where("id = ? AND user_id = ?", checkItemId, userId).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		}).Error
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

// ListCheckItemsSync 增量同步任务检查项列表
// 包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
func (repo *TaskRepoImpl) ListCheckItemsSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.TaskCheckItem, error) {
	scopes := []func(db *gorm.DB) *gorm.DB{
		query.ByKeysetCursor(cursor, cursorID),
		query.SyncOrder(),
	}
	tx := repo.db.WithContext(ctx).Unscoped().
		Model(&models.TaskCheckItem{}).
		Where("user_id = ?", userId).
		Scopes(scopes...)

	if limit <= 0 {
		limit = 100
	}

	var itemModels []*models.TaskCheckItem
	tx = tx.Limit(limit).Find(&itemModels)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCheckItemModels2Entities(itemModels), nil
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

// UpsertComment 幂等写入任务评论：客户端指定 id 时创建或覆盖
// 语义与 Task.Upsert 一致（LWW + create 冲突检测）
func (repo *TaskRepoImpl) UpsertComment(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskComment,
) (*entities.TaskComment, bool, error) {
	if vo.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, err := repo.CreateComment(ctx, userId, vo)
		return entity, true, err
	}
	var existing models.TaskComment
	err := repo.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, createErr := repo.CreateComment(ctx, userId, vo)
		return entity, true, createErr
	}
	if err != nil {
		return nil, false, err
	}
	existingEntity := TaskCommentModel2Entity(&existing)
	outcome, err := types.DecideUpsert(
		existing.CreatedAt, existingEntity.UpdatedAt,
		vo.CreatedAt, vo.UpdatedAt,
		time.Minute,
	)
	if err != nil {
		return nil, false, err
	}
	if outcome == types.UpsertNoop {
		return existingEntity, false, nil
	}
	updateMap := TaskCommentVOToUpdateMap(vo)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now
	updateMap["updated_at"] = time.Now()
	// 覆盖已软删记录（墓碑）时复活：显式清 deleted_at
	updateMap["deleted_at"] = gorm.Expr("NULL")
	if err := repo.db.WithContext(ctx).Unscoped().
		Model(&models.TaskComment{}).
		Where("id = ? AND user_id = ?", vo.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	var updated models.TaskComment
	if err := repo.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&updated).Error; err != nil {
		return nil, false, err
	}
	return TaskCommentModel2Entity(&updated), false, nil
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
	tx := repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ? AND id = ?", userId, commentId)
	// LWW 乐观锁：请求 updatedAt 早于库中版本时不更新
	if !vo.UpdatedAt.IsZero() {
		tx = tx.Where("updated_at <= ?", vo.UpdatedAt)
	}
	return tx.Updates(UpdateTaskCommentValueObjectToMap(vo)).Error
}

// DeleteComment 删除任务评论
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @return error 错误
func (repo *TaskRepoImpl) DeleteComment(ctx context.Context, userId, commentId int64) error {
	// 软删同时推进 updated_at，保证删除墓碑可被增量拉取发现
	return repo.db.
		WithContext(ctx).
		Model(&models.TaskComment{}).
		Where("user_id = ? AND id = ?", userId, commentId).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		}).Error
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

// ListCommentsSync 增量同步任务评论列表
// 包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
func (repo *TaskRepoImpl) ListCommentsSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.TaskComment, error) {
	scopes := []func(db *gorm.DB) *gorm.DB{
		query.ByKeysetCursor(cursor, cursorID),
		query.SyncOrder(),
	}
	tx := repo.db.WithContext(ctx).Unscoped().
		Model(&models.TaskComment{}).
		Where("user_id = ?", userId).
		Scopes(scopes...)

	if limit <= 0 {
		limit = 100
	}

	var commentModels []*models.TaskComment
	tx = tx.Limit(limit).Find(&commentModels)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return TaskCommentModels2Entities(commentModels), nil
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
// 用于项目删除时的级联操作
// 内部走实体 Delete() 维护状态语义，与单条删除路径一致
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) SoftDeleteByProjectId(
	ctx context.Context, userId int64, projectId int64,
) error {
	db := dbs.DBFrom(ctx, repo.db)
	var taskModels []models.Task
	if err := db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Find(&taskModels).Error; err != nil {
		return err
	}
	if len(taskModels) == 0 {
		return nil
	}
	for i := range taskModels {
		taskModels[i].DeletedAt = gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		}
		// 显式推进 updated_at：Save 对非零 UpdatedAt 不覆盖，需手动置位以保证墓碑可增量发现
		taskModels[i].UpdatedAt = time.Now()
	}
	return db.WithContext(ctx).Save(&taskModels).Error
}

// RestoreByProjectId 恢复指定项目下的所有任务
// 用于项目恢复时的级联操作
// 内部逐条加载实体并恢复 deleted_at，与单条恢复路径一致
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) RestoreByProjectId(
	ctx context.Context, userId int64, projectId int64,
) error {
	db := dbs.DBFrom(ctx, repo.db)
	var taskModels []models.Task
	if err := db.WithContext(ctx).
		Unscoped().
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Find(&taskModels).Error; err != nil {
		return err
	}
	if len(taskModels) == 0 {
		return nil
	}
	for i := range taskModels {
		taskModels[i].DeletedAt = gorm.DeletedAt{Valid: false}
		// 显式推进 updated_at，保证恢复事件可增量发现
		taskModels[i].UpdatedAt = time.Now()
	}
	return db.WithContext(ctx).Unscoped().Save(&taskModels).Error
}

// RemoveTagFromTasks 从所有任务中移除指定标签引用
// 用于标签删除时的级联清理：粗筛 tags JSON 数组命中后，Go 侧按字符串精确比对过滤，
// 仅更新实际受影响的记录，并显式推进 updated_at 保证清理结果可被增量同步发现
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return error 错误
func (repo *TaskRepoImpl) RemoveTagFromTasks(
	ctx context.Context, userId int64, tagId int64,
) error {
	db := dbs.DBFrom(ctx, repo.db)
	tagIdStr := strconv.FormatInt(tagId, 10)
	var taskModels []models.Task
	if err := db.WithContext(ctx).
		Where("user_id = ? AND tags LIKE ?", userId, `%"`+tagIdStr+`"%`).
		Find(&taskModels).Error; err != nil {
		return err
	}
	if len(taskModels) == 0 {
		return nil
	}
	// 仅更新 tags 与 updated_at 两列，避免 Save 全量写回覆盖并发修改的其他字段（如 Name/State）
	// Tags 为 serializer:json 列，UpdateColumns 需手动序列化（与 TaskCheckItemVOToUpdateMap 约定一致）
	now := time.Now()
	for i := range taskModels {
		filtered := make([]string, 0, len(taskModels[i].Tags))
		itemChanged := false
		for _, id := range taskModels[i].Tags {
			if id == tagIdStr {
				itemChanged = true
				continue
			}
			filtered = append(filtered, id)
		}
		if !itemChanged {
			continue
		}
		tagsJSON, err := json.Marshal(filtered)
		if err != nil {
			return err
		}
		// 显式推进 updated_at，保证清理事件可增量发现
		if err := db.WithContext(ctx).Model(&models.Task{}).
			Where("id = ?", taskModels[i].ID).
			UpdateColumns(map[string]any{
				"Tags":      string(tagsJSON),
				"UpdatedAt": now,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

// ArchiveByProjectId 归档指定项目下的所有任务
// 用于项目归档时的级联操作
// 内部走实体 Archive() 维护状态语义，与单条归档路径一致
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) ArchiveByProjectId(
	ctx context.Context, userId int64, projectId int64,
) error {
	db := dbs.DBFrom(ctx, repo.db)
	var taskModels []models.Task
	if err := db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Find(&taskModels).Error; err != nil {
		return err
	}
	if len(taskModels) == 0 {
		return nil
	}
	for i := range taskModels {
		entity := TaskModel2Entity(&taskModels[i])
		entity.Archive(nil)
		taskModels[i].ArchivedAt = sql.NullTime{
			Time:  entity.ArchivedAt.Time,
			Valid: true,
		}
	}
	return db.WithContext(ctx).Save(&taskModels).Error
}

// UnarchiveByProjectId 取消归档指定项目下的所有任务
// 用于项目取消归档时的级联操作
// 内部走实体 Unarchive() 维护状态语义，与单条取消归档路径一致
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (repo *TaskRepoImpl) UnarchiveByProjectId(
	ctx context.Context, userId int64, projectId int64,
) error {
	db := dbs.DBFrom(ctx, repo.db)
	var taskModels []models.Task
	if err := db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userId, projectId).
		Find(&taskModels).Error; err != nil {
		return err
	}
	if len(taskModels) == 0 {
		return nil
	}
	for i := range taskModels {
		entity := TaskModel2Entity(&taskModels[i])
		entity.Unarchive()
		taskModels[i].ArchivedAt = sql.NullTime{
			Time:  entity.ArchivedAt.Time,
			Valid: !entity.ArchivedAt.IsNull,
		}
	}
	return db.WithContext(ctx).Save(&taskModels).Error
}
