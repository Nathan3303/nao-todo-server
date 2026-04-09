package event

import (
	"context"
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/repositories"
	"naotodoserver/domain/event/valueobjects"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type EventRepoImpl struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) repositories.Event {
	return &EventRepoImpl{db: db}
}

// GetById 获取事件详情
// @param ctx 上下文
// @param userId 用户ID
// @param eventId 事件ID
// @return 事件详情
// @return error 错误
func (eventRepo *EventRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	eventId int64,
) (*entities.Event, error) {
	// 1. 构建查询条件
	whereCond := &models.Event{}
	whereCond.ID = eventId
	whereCond.UserId = userId
	// 2. 查询数据库
	var event models.Event
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).
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
func (eventRepo *EventRepoImpl) Create(
	ctx context.Context,
	userId int64,
	createEventValueObject *valueobjects.CreateEvent,
) (*entities.Event, error) {
	// 1. 转换为模型
	eventModel := CreateEventValueObjectToModel(createEventValueObject)
	// 2. 插入数据库
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).Create(&eventModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return EventModel2Entity(eventModel), nil
}

// Update 更新事件
// @param ctx 上下文
// @param userId 用户ID
// @param eventId 事件ID
// @param updateEventValueObject 更新事件值对象
// @return error 错误
func (eventRepo *EventRepoImpl) Update(
	ctx context.Context,
	userId int64,
	eventId int64,
	updateEventValueObject *valueobjects.UpdateEvent,
) error {
	// 1. 转换为模型
	var whereCond models.Event
	whereCond.ID = eventId
	whereCond.UserId = userId
	updateCond := UpdateEventValueObjectToModel(updateEventValueObject)
	// 2. 更新数据库
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Updates(updateCond)
	return tx.Error
}

// Delete 删除事件
// @param ctx 上下文
// @param userId 用户ID
// @param eventId 事件ID
// @return error 错误
func (eventRepo *EventRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	eventId int64,
) error {
	// 1. 转换为模型
	var whereCond models.Event
	whereCond.UserId = userId
	whereCond.ID = eventId
	// 2. 删除数据库
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).
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
func (eventRepo *EventRepoImpl) Get(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Event, error) {
	// 1. 转换为模型
	var whereCond models.Event
	whereCond.UserId = userId
	whereCond.TaskId = taskId
	// 2. 查询数据库
	var events []*models.Event
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).
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
func (eventRepo *EventRepoImpl) GetMaxSortId(
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
	eventRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Pluck("MAX(sort_id)", &maxSortId)
	// 3. 返回最大排序 ID
	return maxSortId
}
