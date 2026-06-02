package service

import (
	"context"
	"naotodoserver/domain/checkitem/entities"
	"naotodoserver/domain/checkitem/repositories"
	"naotodoserver/domain/checkitem/valueobjects"
)

// EventDomain 检查事项领域接口
type CheckItemDomain interface {
	// GetById 根据用户 ID和检查事项 ID获取检查事项
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param checkItemId 检查事项 ID
	// @return *entities.CheckItem 检查事项实体
	// @return error 错误信息
	GetById(ctx context.Context, userId int64, checkItemId int64) (*entities.CheckItem, error)

	// Create 创建检查事项
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param createEventValueObject 创建检查事项值对象
	// @return *entities.CheckItem 检查事项实体
	// @return error 错误信息
	Create(
		ctx context.Context,
		userId int64,
		createEventValueObject *valueobjects.CreateCheckItem,
	) (*entities.CheckItem, error)

	// Update 更新检查事项
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param checkItemId 检查事项 ID
	// @param updateEventValueObject 更新检查事项值对象
	// @return error 错误信息
	Update(
		ctx context.Context,
		userId int64,
		checkItemId int64,
		updateEventValueObject *valueobjects.UpdateCheckItem,
	) error

	// Delete 删除检查事项
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param checkItemId 检查事项 ID
	// @return error 错误信息
	Delete(ctx context.Context, userId int64, checkItemId int64) error

	// List 获取检查事项列表
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param taskId 任务 ID
	// @return []*entities.CheckItem 检查事项列表
	// @return error 错误信息
	List(ctx context.Context, userId int64, taskId int64) ([]*entities.CheckItem, error)

	// BatchUpdate 批量更新检查事项
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param batchUpdateEvents 批量更新检查事项值对象集合
	// @return []*entities.CheckItem 更新后的检查事项实体列表
	// @return error 错误信息
	BatchUpdate(
		ctx context.Context,
		userId int64,
		batchUpdateEvents []*valueobjects.BatchUpdateCheckItem,
	) ([]*entities.CheckItem, error)
}

// EventDomainImpl 检查事项领域实现实现
type CheckItemDomainImpl struct {
	checkitemRepo repositories.CheckItem
}
