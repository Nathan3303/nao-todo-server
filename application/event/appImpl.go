package event

import (
	"context"
	"errors"
	"naotodoserver/domain/event/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// RegistDomainImpl 注册检查事项领域服务
func RegistDomainImpl(eventDomain service.EventDomain) EventApp {
	once.Do(func() {
		App = &EventAppImpl{eventDomain: eventDomain}
	})
	return App
}

// GetEventById 获取事件详情
func (eventApp *EventAppImpl) GetEventById(
	ctx context.Context,
	eventId string,
) (*types.EventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(eventId, 10, 64)
	if err != nil {
		return nil, errors.New("检查事项 ID 格式错误")
	}
	// 3. 获取检查事项详情
	eventEntity, err := eventApp.eventDomain.GetById(ctx, userId, eventId64)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntity2Res(eventEntity), nil
}

// CreateEvent 创建事件
func (eventApp *EventAppImpl) CreateEvent(
	ctx context.Context,
	req types.CreateEventReq,
) (*types.EventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体为实体
	createEntity := CreateEventReq2Entity(&req)
	if err := createEntity.IsValid(); err != nil {
		return nil, err
	}
	// 3. 创建检查事项
	eventEntity, err := eventApp.eventDomain.Create(ctx, userId, createEntity)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntity2Res(eventEntity), nil
}

// UpdateEvent 更新事件
func (eventApp *EventAppImpl) UpdateEvent(
	ctx context.Context,
	eventId string,
	req types.UpdateEventReq,
) (*types.UpdateEventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(eventId, 10, 64)
	if err != nil {
		return nil, errors.New("检查事项 ID 格式错误")
	}
	// 3. 转换请求体为实体
	updateEntity := UpdateEventReq2Entity(&req)
	// if err := updateEntity.IsValid(); err != nil {
	// 	return nil, err
	// }
	// 4. 更新检查事项
	err = eventApp.eventDomain.Update(ctx, userId, eventId64, updateEntity)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return &types.UpdateEventRes{EventId: eventId}, nil
}

// DeleteEvent 删除事件
func (eventApp *EventAppImpl) DeleteEvent(
	ctx context.Context,
	eventId string,
) (*types.DeleteEventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(eventId, 10, 64)
	if err != nil {
		return nil, errors.New("检查事项 ID 格式错误")
	}
	// 3. 删除检查事项
	err = eventApp.eventDomain.Delete(ctx, userId, eventId64)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return &types.DeleteEventRes{EventId: eventId}, nil
}

// ListEvent 获取事件列表
func (eventApp *EventAppImpl) ListEvent(
	ctx context.Context,
	todoId string,
) (types.ListEventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	todoId64, err := strconv.ParseInt(todoId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 格式错误")
	}
	// 2. 获取事件列表
	eventEntities, err := eventApp.eventDomain.List(ctx, userId, todoId64)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntities2Reses(eventEntities), nil
}

// ResortEvent 重新排序两个检查事项
func (eventApp *EventAppImpl) ResortEvents(
	ctx context.Context,
	req types.ResortEventsReq,
) (*types.ResortEventsRes, error) {
	panic("unimplement")
}
