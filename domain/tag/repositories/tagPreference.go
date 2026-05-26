package repositories

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/valueobjects"
)

// 标签偏好接口
type TagPreference interface {
	// 根据用户和标签ID获取标签偏好
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagId 标签ID
	// @return entities.TagPreference 标签偏好实体
	// @return error 错误
	Get(ctx context.Context, userId int64, tagId int64) (*entities.TagPreference, error)

	// 保存标签偏好
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagId 标签ID
	// @param saveTagPreferenceValueObject 保存标签偏好值对象
	// @return error 错误
	Save(
		ctx context.Context,
		userId int64,
		tagId int64,
		saveTagPreferenceValueObject *valueobjects.SaveTagPreference,
	) error

	// 删除标签偏好
	// @param ctx 上下文
	// @param userId 用户ID
	// @param tagId 标签ID
	// @return error 错误
	Delete(ctx context.Context, userId int64, tagId int64) error
}
