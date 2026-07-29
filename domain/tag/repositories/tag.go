package repositories

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/valueobjects"
)

// TagRepository 标签仓库接口
type TagRepository interface {
	// 根据用户获取标签
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagId 标签ID
	// @return entities.Tag 标签实体
	// @return error 错误
	GetById(ctx context.Context, userId int64, tagId int64) (*entities.Tag, error)

	// 创建标签
	// @param ctx 上下文
	// @param userId 用户ID
	// @param createTagValueObject 创建标签值对象
	// @return entities.Tag 标签实体
	// @return error 错误
	Create(
		ctx context.Context,
		userId int64,
		createTagValueObject *valueobjects.CreateTag,
	) (*entities.Tag, error)

	// 更新标签
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagId 标签ID
	// @param updateTagValueObject 更新标签值对象
	// @return error 错误
	Update(
		ctx context.Context,
		userId int64,
		tagId int64,
		updateTagValueObject *valueobjects.UpdateTag,
	) error

	// 删除标签
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagId 标签ID
	// @return error 错误
	Delete(ctx context.Context, userId int64, tagId int64) error

	// 获取标签列表
	// @param ctx 上下文
	// @param userId 用户ID
	// @return []*entities.Tag 标签实体列表
	// @return error 错误
	Get(ctx context.Context, userId int64) ([]*entities.Tag, error)

	// 根据标签ID列表获取标签列表
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagIds 标签ID列表
	// @return []*entities.Tag 标签实体列表
	// @return error 错误
	GetByIds(ctx context.Context, userId int64, tagIds []int64) ([]*entities.Tag, error)

	// 获取最大排序 ID
	GetMaxSortId(ctx context.Context, userId int64) uint16

	// 批量更新标签
	BatchUpdate(
		ctx context.Context,
		userId int64,
		batchUpdateTags []*valueobjects.BatchUpdateTag,
	) ([]*entities.Tag, error)
}
