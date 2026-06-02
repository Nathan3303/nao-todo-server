package checkitem

import (
	"context"
	"errors"
	"naotodoserver/domain/checkitem/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// NewCheckItemApp 创建检查事项应用层实例
func NewCheckItemApp(checkitemDomain service.CheckItemDomain) CheckItemApp {
	impl := &CheckItemAppImpl{checkitemDomain: checkitemDomain}
	return impl
}

// GetCheckItemById 获取检查事项详情
// @param ctx 上下文
// @param itemId 检查事项 ID
// @return 检查事项详情
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) GetCheckItemById(
	ctx context.Context,
	itemId string,
) (*types.GetCheckItemRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		return nil, errors.New("检查事项 ID 格式错误")
	}
	// 3. 获取检查事项详情
	e, err := checkitemApp.checkitemDomain.GetById(ctx, userId, eventId64)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntityToGetRes(e), nil
}

// CreateCheckItem 创建检查事项
// @param ctx 上下文
// @param createReq 创建检查事项请求体
// @return 创建检查事项详情
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) CreateCheckItem(
	ctx context.Context,
	createReq *types.CreateCheckItemReq,
) (*types.CreateCheckItemRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体为实体
	createVO, err := CreateEventReqToValueObject(userId, createReq)
	if err != nil {
		return nil, err
	}
	// 3. 创建检查事项
	e, err := checkitemApp.checkitemDomain.Create(
		ctx,
		userId,
		createVO,
	)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntityToCreateRes(e), nil
}

// UpdateCheckItem 更新检查事项
// @param ctx 上下文
// @param itemId 检查事项 ID
// @param updateReq 更新检查事项请求体
// @return 更新检查事项详情
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) UpdateCheckItem(
	ctx context.Context,
	itemId string,
	updateReq *types.UpdateCheckItemReq,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		return errors.New("检查事项 ID 格式错误")
	}
	// 3. 转换请求体为实体
	updateVO, err := UpdateEventReqToValueObject(updateReq)
	if err != nil {
		return err
	}
	// 4. 更新检查事项
	return checkitemApp.checkitemDomain.Update(
		ctx,
		userId,
		eventId64,
		updateVO,
	)
}

// DeleteCheckItem 删除检查事项
// @param ctx 上下文
// @param itemId 检查事项 ID
// @return 删除检查事项详情
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) DeleteCheckItem(
	ctx context.Context,
	itemId string,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换检查事项 ID
	eventId64, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		return errors.New("检查事项 ID 格式错误")
	}
	// 3. 删除检查事项
	return checkitemApp.checkitemDomain.Delete(ctx, userId, eventId64)
}

// ListCheckItem 获取检查事项列表
// @param ctx 上下文
// @param taskId 待办任务 ID
// @return 检查事项列表
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) ListCheckItem(
	ctx context.Context,
	taskId string,
) (types.ListCheckItemRes, error) {
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
	items, err := checkitemApp.checkitemDomain.List(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	// 返回结果
	return EventEntities2Reses(items), nil
}

// ResortEvent 重新排序两个检查事项
// @param ctx 上下文
// @param resortReq 排序检查事项请求体
// @return 排序检查事项详情
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) ResortCheckItems(
	ctx context.Context,
	resortReq *types.ResortCheckItemsReq,
) (*types.ResortCheckItemsRes, error) {
	panic("unimplement")
}

// BatchUpdateCheckItems 批量更新检查事项
// @param ctx 上下文
// @param batchReq 批量更新检查事项请求体
// @return 批量更新检查事项响应
// @return error 错误信息
func (checkitemApp *CheckItemAppImpl) BatchUpdateCheckItems(
	ctx context.Context,
	batchReq *types.BatchUpdateCheckItemReq,
) (*types.BatchUpdateCheckItemRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体为值对象
	batchVOs, err := BatchUpdateEventReqToValueObjects(batchReq)
	if err != nil {
		return nil, err
	}
	// 3. 调用领域服务批量更新事件
	updatedItems, err := checkitemApp.checkitemDomain.BatchUpdate(ctx, userId, batchVOs)
	if err != nil {
		return nil, err
	}
	// 4. 转换实体为响应
	resList := EventEntities2Reses(updatedItems)
	// 返回结果
	return &types.BatchUpdateCheckItemRes{
		UpdatedCount: int64(len(resList)),
		Events:       resList,
	}, nil
}
