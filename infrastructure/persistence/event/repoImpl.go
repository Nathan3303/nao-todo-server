package event

import (
	"context"
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/repositories"
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
func (eventRepo *EventRepoImpl) Create(
	ctx context.Context,
	createEntity *entities.Event,
) (*entities.Event, error) {
	// 1. 转换为模型
	eventModel := EventEntity2Model(createEntity)
	// 2. 插入数据库
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).Create(&eventModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return EventModel2Entity(eventModel), nil
}

// Update 更新事件
func (eventRepo *EventRepoImpl) Update(
	ctx context.Context,
	whereEntity *entities.Event,
	updateEntity *entities.Event,
) error {
	// 1. 转换为模型
	whereCond := EventEntity2Model(whereEntity)
	updateCond := EventEntity2Model(updateEntity)
	// 2. 更新数据库
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Updates(updateCond)
	return tx.Error
}

// Delete 删除事件
func (eventRepo *EventRepoImpl) Delete(
	ctx context.Context,
	whereEntity *entities.Event,
) error {
	// 1. 转换为模型
	whereCond := EventEntity2Model(whereEntity)
	// 2. 删除数据库
	tx := eventRepo.db.WithContext(ctx).Model(&models.Event{}).
		Where(whereCond).
		Delete(&models.Event{})
	return tx.Error
}

// Get 获取事件列表
func (eventRepo *EventRepoImpl) Get(
	ctx context.Context,
	whereEntity *entities.Event,
) ([]*entities.Event, error) {
	// 1. 转换为模型
	whereCond := EventEntity2Model(whereEntity)
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
