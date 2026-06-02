package service

import (
	"context"
	"naotodoserver/domain/checkitem/entities"
	"naotodoserver/domain/checkitem/repositories"
	"naotodoserver/domain/checkitem/valueobjects"
)

// NewCheckItemDomain 创建检查事项领域服务
func NewCheckItemDomain(checkitemRepo repositories.CheckItem) CheckItemDomain {
	return &CheckItemDomainImpl{checkitemRepo: checkitemRepo}
}

// GetById 根据用户 ID 和事件 ID 获取检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param checkItemId 事件 ID
// @return *entities.CheckItem 事件实体
// @return error 错误信息
func (checkitemDomain *CheckItemDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	checkItemId int64,
) (*entities.CheckItem, error) {
	return checkitemDomain.checkitemRepo.GetById(ctx, userId, checkItemId)
}

// Create 创建检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param createEventValueObject 创建检查事项值对象
// @return *entities.CheckItem 检查事项实体
// @return error 错误信息
func (checkitemDomain *CheckItemDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEventValueObject *valueobjects.CreateCheckItem,
) (*entities.CheckItem, error) {
	// 查找最大排序 ID 并加 1
	// 确保排序 ID 永远比上一次的排序 ID 大
	// 如果没有检查事项，排序 ID 为 256
	createEventValueObject.SortId = checkitemDomain.checkitemRepo.GetMaxSortId(
		ctx,
		userId,
		createEventValueObject.TaskId,
	) + 1
	// 创建检查事项
	return checkitemDomain.checkitemRepo.Create(ctx, userId, createEventValueObject)
}

// Update 更新检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param checkItemId 检查事项 ID
// @param updateEventValueObject 更新检查事项值对象
// @return error 错误信息
func (checkitemDomain *CheckItemDomainImpl) Update(
	ctx context.Context,
	userId int64,
	checkItemId int64,
	updateEventValueObject *valueobjects.UpdateCheckItem,
) error {
	return checkitemDomain.checkitemRepo.Update(ctx, userId, checkItemId, updateEventValueObject)
}

// Delete 删除检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param checkItemId 检查事项 ID
// @return error 错误信息
func (checkitemDomain *CheckItemDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	checkItemId int64,
) error {
	return checkitemDomain.checkitemRepo.Delete(ctx, userId, checkItemId)
}

// List 根据用户 ID 和任务 ID 获取检查事项列表
// @param ctx 上下文
// @param userId 用户 ID
// @param taskId 任务 ID
// @return []*entities.CheckItem 检查事项列表
// @return error 错误信息
func (checkitemDomain *CheckItemDomainImpl) List(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.CheckItem, error) {
	return checkitemDomain.checkitemRepo.Get(ctx, userId, taskId)
}

// BatchUpdate 批量更新检查事项
// @param ctx 上下文
// @param userId 用户 ID
// @param batchUpdateEvents 批量更新检查事项值对象集合
// @return []*entities.CheckItem 更新后的检查事项实体列表
// @return error 错误信息
func (checkitemDomain *CheckItemDomainImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateEvents []*valueobjects.BatchUpdateCheckItem,
) ([]*entities.CheckItem, error) {
	return checkitemDomain.checkitemRepo.BatchUpdate(ctx, userId, batchUpdateEvents)
}
