package event

import (
	"context"
	"naotodoserver/domain/checkitem/entities"
	"naotodoserver/domain/checkitem/repositories"
	"naotodoserver/domain/checkitem/valueobjects"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type CheckItemRepoImpl struct {
	db *gorm.DB
}

func NewCheckItemRepo(db *gorm.DB) repositories.CheckItem {
	return &CheckItemRepoImpl{db: db}
}

// GetById 获取事件详情
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 事件ID
// @return 事件详情
// @return error 错误
func (checkitemRepo *CheckItemRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	checkItemId int64,
) (*entities.CheckItem, error) {
	// 1. 构建查询条件
	whereCond := &models.Event{}
	whereCond.ID = checkItemId
	whereCond.UserId = userId
	// 2. 查询数据库
	var event models.Event
	tx := checkitemRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		First(&event)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return EventModel2Entity(&event), nil
}

// Create 创建事件
// @param ctx 上下文
// @param createEntity 创建事件实体
// @return 事件详情
// @return error 错误
func (checkitemRepo *CheckItemRepoImpl) Create(
	ctx context.Context,
	userId int64,
	createEventValueObject *valueobjects.CreateCheckItem,
) (*entities.CheckItem, error) {
	// 1. 转换为模型
	eventModel := CreateEventValueObjectToModel(createEventValueObject)
	// 2. 插入数据库
	tx := checkitemRepo.db.WithContext(ctx).Model(&models.Event{}).Create(&eventModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return EventModel2Entity(eventModel), nil
}

// Update 更新事件
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 事件ID
// @param updateEventValueObject 更新事件值对象
// @return error 错误
func (checkitemRepo *CheckItemRepoImpl) Update(
	ctx context.Context,
	userId int64,
	checkItemId int64,
	updateEventValueObject *valueobjects.UpdateCheckItem,
) error {
	// 1. 转换为模型
	var whereCond models.Event
	whereCond.ID = checkItemId
	whereCond.UserId = userId
	updateCond := UpdateEventValueObjectToMap(updateEventValueObject)
	// 2. 更新数据库
	tx := checkitemRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Updates(updateCond)
	return tx.Error
}

// Delete 删除事件
// @param ctx 上下文
// @param userId 用户ID
// @param checkItemId 事件ID
// @return error 错误
func (checkitemRepo *CheckItemRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	checkItemId int64,
) error {
	// 1. 转换为模型
	var whereCond models.Event
	whereCond.UserId = userId
	whereCond.ID = checkItemId
	// 2. 删除数据库
	tx := checkitemRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Delete(&models.Event{})
	return tx.Error
}

// Get 获取事件列表
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return 事件列表
// @return error 错误
func (checkitemRepo *CheckItemRepoImpl) Get(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.CheckItem, error) {
	// 1. 转换为模型
	var whereCond models.Event
	whereCond.UserId = userId
	whereCond.TaskId = taskId
	// 2. 查询数据库
	var events []*models.Event
	tx := checkitemRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Find(&events)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return EventModels2Entities(events), nil
}

// GetMaxSortId 获取最大排序 ID
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return maxSortId 最大排序 ID
// @return error 错误信息
func (checkitemRepo *CheckItemRepoImpl) GetMaxSortId(
	ctx context.Context,
	userId int64,
	taskId int64,
) uint16 {
	// 1. 构建查询条件
	var whereCond models.Event
	whereCond.UserId = userId
	whereCond.TaskId = taskId
	// 2. 查询数据库
	var maxSortId uint16 = 255
	checkitemRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Pluck("MAX(sort_id)", &maxSortId)
	// 3. 返回最大排序 ID
	return maxSortId
}

// BatchUpdate 批量更新事件
// @param ctx 上下文
// @param userId 用户ID
// @param batchUpdateEvents 批量更新事件值对象集合
// @return []*entities.CheckItem 更新后的事件实体列表
// @return error 错误
func (checkitemRepo *CheckItemRepoImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateEvents []*valueobjects.BatchUpdateCheckItem,
) ([]*entities.CheckItem, error) {
	// 开始事务
	tx := checkitemRepo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	updatedEventIds := make([]int64, 0, len(batchUpdateEvents))

	// 逐个更新事件
	for _, batchEvent := range batchUpdateEvents {
		whereCond := &models.Event{}
		whereCond.ID = batchEvent.Id
		whereCond.UserId = userId
		updateCond := BatchUpdateEventValueObjectToMap(batchEvent)

		// 更新数据库
		if err := tx.Model(&models.Event{}).
			Where(whereCond).
			Updates(updateCond).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		updatedEventIds = append(updatedEventIds, batchEvent.Id)
	}

	// 查询更新后的事件
	var updatedEvents []*models.Event
	if err := tx.Model(&models.Event{}).
		Where("id IN ? AND user_id = ?", updatedEventIds, userId).
		Find(&updatedEvents).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return EventModels2Entities(updatedEvents), nil
}
