package repositories

import (
	"context"
	"naotodoserver/domain/checkitem/entities"
	"naotodoserver/domain/checkitem/valueobjects"
)

// 检查事项仓库接口
type CheckItem interface {
	// GetById 根据用户 ID 和事件 ID 获取事件
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param checkItemId 事件 ID
	// @return *entities.CheckItem 事件实体
	// @return error 错误信息
	GetById(ctx context.Context, userId int64, checkItemId int64) (*entities.CheckItem, error)

	// Create 创建事件
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param createEventValueObject 创建事件值对象
	// @return *entities.CheckItem 事件实体
	// @return error 错误信息
	Create(
		ctx context.Context,
		userId int64,
		createEventValueObject *valueobjects.CreateCheckItem,
	) (*entities.CheckItem, error)

	// Update 更新事件
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param checkItemId 事件 ID
	// @param updateEventValueObject 更新事件值对象
	// @return error 错误信息
	Update(
		ctx context.Context,
		userId int64,
		checkItemId int64,
		updateEventValueObject *valueobjects.UpdateCheckItem,
	) error

	// Delete 删除事件
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param checkItemId 事件 ID
	// @return error 错误信息
	Delete(ctx context.Context, userId int64, checkItemId int64) error

	// Get 获取事件列表
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param taskId 任务 ID
	// @return []*entities.CheckItem 事件列表
	// @return error 错误信息
	Get(ctx context.Context, userId int64, taskId int64) ([]*entities.CheckItem, error)

	// GetMaxSortId 获取最大排序 ID
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param taskId 任务 ID
	// @return maxSortId 最大排序 ID
	GetMaxSortId(ctx context.Context, userId int64, taskId int64) uint16

	// BatchUpdate 批量更新事件
	// @param ctx 上下文
	// @param userId 用户 ID
	// @param batchUpdateEvents 批量更新事件值对象集合
	// @return []*entities.CheckItem 更新后的事件实体列表
	// @return error 错误信息
	BatchUpdate(
		ctx context.Context,
		userId int64,
		batchUpdateEvents []*valueobjects.BatchUpdateCheckItem,
	) ([]*entities.CheckItem, error)
}
