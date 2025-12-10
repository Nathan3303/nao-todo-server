package task

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/vo"
	"naotodoserver/infrastructure/persistence/models"

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
