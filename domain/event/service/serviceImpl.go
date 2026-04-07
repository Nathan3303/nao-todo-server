package service

import (
	"context"
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/repositories"
	"naotodoserver/domain/event/valueobjects"
)

// NewEventDomain 创建检查事项领域服务
func NewEventDomain(eventRepo repositories.Event) EventDomain {
	return &EventDomainImpl{eventRepo: eventRepo}
}

// GetById 根据用户 ID 和事件 ID 获取事件
// @param ctx 上下文
// @param userId 用户 ID
// @param eventId 事件 ID
// @return *entities.Event 事件实体
// @return error 错误信息
func (eventDomain *EventDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	eventId int64,
) (*entities.Event, error) {
	return eventDomain.eventRepo.GetById(ctx, userId, eventId)
}

// Create 创建事件
// @param ctx 上下文
// @param userId 用户 ID
// @param createEventValueObject 创建事件值对象
// @return *entities.Event 事件实体
// @return error 错误信息
func (eventDomain *EventDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEventValueObject *valueobjects.CreateEvent,
) (*entities.Event, error) {
	return eventDomain.eventRepo.Create(ctx, userId, createEventValueObject)
}

// Update 更新事件
// @param ctx 上下文
// @param userId 用户 ID
// @param eventId 事件 ID
// @param updateEventValueObject 更新事件值对象
// @return error 错误信息
func (eventDomain *EventDomainImpl) Update(
	ctx context.Context,
	userId int64,
	eventId int64,
	updateEventValueObject *valueobjects.UpdateEvent,
) error {
	return eventDomain.eventRepo.Update(ctx, userId, eventId, updateEventValueObject)
}

// Delete 删除事件
// @param ctx 上下文
// @param userId 用户 ID
// @param eventId 事件 ID
// @return error 错误信息
func (eventDomain *EventDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	eventId int64,
) error {
	return eventDomain.eventRepo.Delete(ctx, userId, eventId)
}

// List 根据用户 ID 和任务 ID 获取事件列表
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务 ID
// @return []*entities.Event 事件列表
// @return error 错误信息
func (eventDomain *EventDomainImpl) List(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Event, error) {
	return eventDomain.eventRepo.Get(ctx, userId, taskId)
}
