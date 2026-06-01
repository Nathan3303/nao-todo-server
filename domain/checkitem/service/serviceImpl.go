package service

import (
	"context"
	"naotodoserver/domain/checkitem/entities"
	"naotodoserver/domain/checkitem/repositories"
	"naotodoserver/domain/checkitem/valueobjects"
)

// NewEventDomain 创建检查事项领域服务
func NewEventDomain(eventRepo repositories.Event) EventDomain {
	return &EventDomainImpl{eventRepo: eventRepo}
}

// GetById 根据用户 ID 和事件 ID 获取检查事项
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

// Create 创建检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param createEventValueObject 创建检查事项值对象
// @return *entities.Event 检查事项实体
// @return error 错误信息
func (eventDomain *EventDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEventValueObject *valueobjects.CreateEvent,
) (*entities.Event, error) {
	// 查找最大排序 ID 并加 1
	// 确保排序 ID 永远比上一次的排序 ID 大
	// 如果没有检查事项，排序 ID 为 256
	createEventValueObject.SortId = eventDomain.eventRepo.GetMaxSortId(
		ctx,
		userId,
		createEventValueObject.TaskId,
	) + 1
	// 创建检查事项
	return eventDomain.eventRepo.Create(ctx, userId, createEventValueObject)
}

// Update 更新检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param eventId 检查事项 ID
// @param updateEventValueObject 更新检查事项值对象
// @return error 错误信息
func (eventDomain *EventDomainImpl) Update(
	ctx context.Context,
	userId int64,
	eventId int64,
	updateEventValueObject *valueobjects.UpdateEvent,
) error {
	return eventDomain.eventRepo.Update(ctx, userId, eventId, updateEventValueObject)
}

// Delete 删除检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param eventId 检查事项 ID
// @return error 错误信息
func (eventDomain *EventDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	eventId int64,
) error {
	return eventDomain.eventRepo.Delete(ctx, userId, eventId)
}

// List 根据用户 ID 和任务 ID 获取检查事项列表
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务 ID
// @return []*entities.Event 检查事项列表
// @return error 错误信息
func (eventDomain *EventDomainImpl) List(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Event, error) {
	return eventDomain.eventRepo.Get(ctx, userId, taskId)
}

// BatchUpdate 批量更新检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param batchUpdateEvents 批量更新检查事项值对象集合
// @return []*entities.Event 更新后的检查事项实体列表
// @return error 错误信息
func (eventDomain *EventDomainImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateEvents []*valueobjects.BatchUpdateEvent,
) ([]*entities.Event, error) {
	return eventDomain.eventRepo.BatchUpdate(ctx, userId, batchUpdateEvents)
}
