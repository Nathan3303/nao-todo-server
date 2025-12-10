package service

import (
	"context"
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/repositories"
	"time"
)

// NewEventDomain 创建检查事项领域服务
func NewEventDomain(eventRepo repositories.Event) EventDomain {
	return &EventDomainImpl{eventRepo: eventRepo}
}

// GetById 根据用户 ID 和检查事项 ID 获取检查事项
func (eventDomain *EventDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	eventId int64,
) (*entities.Event, error) {
	return eventDomain.eventRepo.GetById(ctx, userId, eventId)
}

// Create 创建检查事项
func (eventDomain *EventDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEntity *entities.Event,
) (*entities.Event, error) {
	createEntity.UserId = userId
	createEntity.SortId = int32(time.Now().UnixNano())
	return eventDomain.eventRepo.Create(ctx, createEntity)
}

// Update 更新检查事项
func (eventDomain *EventDomainImpl) Update(
	ctx context.Context,
	userId int64,
	eventId int64,
	updateEntity *entities.Event,
) error {
	whereEntity := &entities.Event{UserId: userId, Id: eventId}
	return eventDomain.eventRepo.Update(ctx, whereEntity, updateEntity)
}

// Delete 删除检查事项
func (eventDomain *EventDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	eventId int64,
) error {
	whereEntity := &entities.Event{UserId: userId, Id: eventId}
	return eventDomain.eventRepo.Delete(ctx, whereEntity)
}

// List 根据用户 ID 和任务 ID 获取检查事项列表
func (eventDomain *EventDomainImpl) List(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Event, error) {
	whereEntity := &entities.Event{UserId: userId, TaskId: taskId}
	return eventDomain.eventRepo.Get(ctx, whereEntity)
}
