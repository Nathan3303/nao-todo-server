package task

import (
	"context"
	"naotodoserver/consts"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/vo"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

type TaskRepoImpl struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) repositories.Task {
	return &TaskRepoImpl{db: db}
}

/*
 * Get task by id
 * 根据任务ID获取任务信息
 */
func (taskRepo *TaskRepoImpl) GetById(
	ctx context.Context,
	whereEntity *entities.Task,
) (*entities.Task, error) {
	// 1. 创建结果模型
	taskModel := &models.Task{}
	// 2. 转换查询实体到模型
	whereCond := TaskEntity2Model(whereEntity)
	// 3. 查询
	tx := taskRepo.db.WithContext(ctx).Where(whereCond).First(taskModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 转换模型到实体并返回
	return TaskModel2Entity(taskModel), nil
}

/*
 * Create task
 * 创建任务
 */
func (taskRepo *TaskRepoImpl) Create(
	ctx context.Context,
	createEntity *entities.Task,
) (*entities.Task, error) {
	// 1. 转换创建实体到模型
	createModel := TaskEntity2Model(createEntity)
	// 2. 创建
	tx := taskRepo.db.WithContext(ctx).Create(createModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 转换模型到实体并返回
	return TaskModel2Entity(createModel), nil
}

/*
 * Update task
 * 更新任务
 */
func (taskRepo *TaskRepoImpl) Update(
	ctx context.Context,
	whereEntity *entities.Task,
	updateEntity *entities.Task,
) error {
	// 1. 转换查询实体到模型
	whereCond := TaskEntity2Model(whereEntity)
	// 2. 转换更新实体到模型
	updateModel := TaskEntity2Model(updateEntity)
	// 3. 更新
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Updates(updateModel)
	return tx.Error
}

/*
 * Delete task
 * 删除任务
 */
func (taskRepo *TaskRepoImpl) Delete(
	ctx context.Context,
	whereEntity *entities.Task,
) error {
	// 1. 转换查询实体到模型
	whereCond := TaskEntity2Model(whereEntity)
	// 2. 删除
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Delete(&models.Task{})
	return tx.Error
}

/*
 * Restore task
 * 恢复任务
 */
func (taskRepo *TaskRepoImpl) Restore(
	ctx context.Context,
	whereEntity *entities.Task,
) error {
	// 1. 转换查询实体到模型
	whereCond := TaskEntity2Model(whereEntity)
	// 2. 恢复
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).Unscoped().
		Where(whereCond).
		Update("deleted_at", nil)
	return tx.Error
}

/*
 * List task
 * 获取任务列表
 */
func (taskRepo *TaskRepoImpl) List(
	ctx context.Context,
	whereEntity *entities.Task,
	pagination *vo.Pagination,
) ([]*entities.Task, *vo.Pagination, error) {
	// 1. 转换查询实体到模型
	whereCond := TaskEntity2Model(whereEntity)
	// 2. 查询任务总数
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where(whereCond).
		Count(&pagination.Total)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	// 2. 查询任务列表
	taskModels := []*models.Task{}
	tx = tx.Scopes(PaginationVO2Scopes(pagination)).Find(&taskModels)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	// 3. 转换模型到实体并返回
	taskEntities := TaskModels2Entities(taskModels)
	return taskEntities, pagination, nil
}

/*
 * List task with query TX
 * 通过 数据库操作句柄 获取任务列表
 */
func (taskRepo *TaskRepoImpl) ListWithQueryTx(
	ctx context.Context,
	tx *gorm.DB,
	pagination *vo.Pagination,
) ([]*entities.Task, *vo.Pagination, error) {
	// 1. 查询任务总数
	tx = tx.WithContext(ctx).Model(&models.Task{}).
		Count(&pagination.Total)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	// 2. 查询任务列表
	taskModels := []*models.Task{}
	tx = tx.Scopes(PaginationVO2Scopes(pagination)).Find(&taskModels)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	// 3. 转换模型到实体并返回
	taskEntities := TaskModels2Entities(taskModels)
	return taskEntities, pagination, nil
}

/*
 * Build query tx
 * 构建查询操作符
 */
func (taskRepo *TaskRepoImpl) BuildQueryTx(
	ctx context.Context,
	query *vo.TaskQuery,
) (*gorm.DB, error) {
	// 1. 构建指定模型的操作符
	tx := taskRepo.db.WithContext(ctx).Model(&models.Task{}).
		Where("user_id = ?", query.UserId)
	// 2. 处理 Project 或 Tag ID 过滤条件
	if query.ProjectId > 0 {
		tx = tx.Where("project_id = ?", query.ProjectId)
	} else if query.TagId != "" {
		tx = tx.Where("tags LIKE ?", "%"+query.TagId+"%")
	}
	// 3. 处理名称和描述过滤条件
	if query.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.Description != "" {
		tx = tx.Where("description LIKE ?", "%"+query.Description+"%")
	}
	// 4. 处理状态过滤条件
	if query.State != "" {
		var stateIDs []int8
		for s := range strings.SplitSeq(query.State, ",") {
			if val, exists := consts.TodoStateMap[s]; exists {
				stateIDs = append(stateIDs, val)
			}
		}
		if len(stateIDs) > 0 {
			tx = tx.Where("state IN ?", stateIDs)
		}
	}
	// 5. 处理优先级过滤条件
	if query.Priority != "" {
		var priorityIDs []int8
		for _, p := range strings.Split(query.Priority, ",") {
			if val, exists := consts.TodoPriorityMap[p]; exists {
				priorityIDs = append(priorityIDs, val)
			}
		}
		if len(priorityIDs) > 0 {
			tx = tx.Where("priority IN ?", priorityIDs)
		}
	}
	// 6. 处理开始时间和结束时间过滤条件
	if query.StartAt != "" {
		tx = tx.Where("start_at >= ?", query.StartAt)
	}
	if query.EndAt != "" {
		tx = tx.Where("end_at <= ?", query.EndAt)
	}
	// 7. 处理所有布尔类型的过滤条件
	if query.IsDeleted {
		tx = tx.Where("deleted_at IS NOT NULL")
	}
	if query.IsArchived {
		tx = tx.Where("archived_at IS NOT NULL")
	}
	if query.IsStarMarked {
		tx = tx.Where("star_mark_at IS NOT NULL")
	}
	if query.IsGivenUp {
		tx = tx.Where("given_up_at IS NOT NULL")
	}
	// 8. 处理相对日期过滤条件
	if query.RelativeDate != "" {
		switch query.RelativeDate {
		case "today":
			tx.Where("end_at >= ?", time.Now().Format("2006-01-02"))
		case "tomorrow":
			tx.Where("end_at >= ?", time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
		case "week":
			{
				start, end := utils.GetWeekRange(time.Now())
				tx.Where("end_at >= ? and end_at <= ?", start, end)
			}
		case "month":
			tx.Where("end_at >= ?", time.Now().AddDate(0, 0, 7).Format("2006-01-02"))
		case "-today":
			tx.Where("end_at < ?", time.Now().Format("2006-01-02"))
		}
	}
	// 9. 处理排序条件
	if query.Sort != "" {
		var splited = strings.Split(query.Sort, ":")
		if len(splited) == 2 {
			tx.Order(utils.ToSnakeCase(splited[0]) + " " + splited[1])
		}
	}
	// 10. 返回
	return tx, nil
}
