package event

import (
	"context"
	"errors"
	"naotodoserver/domain/checkitem/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// NewEventApp 创建检查事项应用层实例
func NewEventApp(eventDomain service.EventDomain) EventApp {
	impl := &EventAppImpl{eventDomain: eventDomain}
	return impl
}

// GetEventById 获取检查事项详情
// @param ctx 上下文
// @param eventId 检查事项 ID
// @return 检查事项详情
// @return error 错误信息
func (eventApp *EventAppImpl) GetEventById(
	ctx context.Context,
	eventId string,
) (*types.GetEventRes, error) {
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
	return EventEntityToGetRes(eventEntity), nil
}

// CreateEvent 创建检查事项
// @param ctx 上下文
// @param createEventReq 创建检查事项请求体
// @return 创建检查事项详情
// @return error 错误信息
func (eventApp *EventAppImpl) CreateEvent(
	ctx context.Context,
	createEventReq *types.CreateEventReq,
) (*types.CreateEventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体为实体
	createEntity, err := CreateEventReqToValueObject(userId, createEventReq)
	if err != nil {
		return nil, err
	}
	// 3. 创建检查事项
	eventEntity, err := eventApp.eventDomain.Create(
		ctx,
		userId,
		createEntity,
	)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntityToCreateRes(eventEntity), nil
}

// UpdateEvent 更新检查事项
// @param ctx 上下文
// @param eventId 检查事项 ID
// @param updateEventReq 更新检查事项请求体
// @return 更新检查事项详情
// @return error 错误信息
func (eventApp *EventAppImpl) UpdateEvent(
	ctx context.Context,
	eventId string,
	updateEventReq *types.UpdateEventReq,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(eventId, 10, 64)
	if err != nil {
		return errors.New("检查事项 ID 格式错误")
	}
	// 3. 转换请求体为实体
	updateEntity, err := UpdateEventReqToValueObject(updateEventReq)
	if err != nil {
		return err
	}
	// 4. 更新检查事项
	return eventApp.eventDomain.Update(
		ctx,
		userId,
		eventId64,
		updateEntity,
	)
}

// DeleteEvent 删除检查事项
// @param ctx 上下文
// @param eventId 检查事项 ID
// @return 删除检查事项详情
// @return error 错误信息
func (eventApp *EventAppImpl) DeleteEvent(
	ctx context.Context,
	eventId string,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(eventId, 10, 64)
	if err != nil {
		return errors.New("检查事项 ID 格式错误")
	}
	// 3. 删除检查事项
	return eventApp.eventDomain.Delete(ctx, userId, eventId64)
}

// ListEvent 获取检查事项列表
// @param ctx 上下文
// @param taskId 待办任务 ID
// @return 检查事项列表
// @return error 错误信息
func (eventApp *EventAppImpl) ListEvent(
	ctx context.Context,
	taskId string,
) (types.ListEventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 格式错误")
	}
	// 2. 获取事件列表
	eventEntities, err := eventApp.eventDomain.List(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntities2Reses(eventEntities), nil
}

// ResortEvent 重新排序两个检查事项
// @param ctx 上下文
// @param resortEventsReq 排序检查事项请求体
// @return 排序检查事项详情
// @return error 错误信息
func (eventApp *EventAppImpl) ResortEvents(
	ctx context.Context,
	resortEventsReq *types.ResortEventsReq,
) (*types.ResortEventsRes, error) {
	panic("unimplement")
}

// BatchUpdateEvents 批量更新检查事项
// @param ctx 上下文
// @param batchUpdateEventReq 批量更新检查事项请求体
// @return 批量更新检查事项响应
// @return error 错误信息
func (eventApp *EventAppImpl) BatchUpdateEvents(
	ctx context.Context,
	batchUpdateEventReq *types.BatchUpdateEventReq,
) (*types.BatchUpdateEventRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体为值对象
	batchVOs, err := BatchUpdateEventReqToValueObjects(batchUpdateEventReq)
	if err != nil {
		return nil, err
	}
	// 3. 调用领域服务批量更新事件
	updatedEntities, err := eventApp.eventDomain.BatchUpdate(ctx, userId, batchVOs)
	if err != nil {
		return nil, err
	}
	// 4. 转换实体为响应
	eventResList := EventEntities2Reses(updatedEntities)
	// 返回结果
	return &types.BatchUpdateEventRes{
		UpdatedCount: int64(len(eventResList)),
		Events:       eventResList,
	}, nil
}
