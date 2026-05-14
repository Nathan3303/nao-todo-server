package service

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
)

// 标签领域服务接口
type TagDomain interface {
	// 根据 ID 获取标签
	GetById(ctx context.Context, userId int64, tagId int64) (*entities.Tag, error)

	// 创建标签
	Create(
		ctx context.Context,
		userId int64,
		createTagValueObject *valueobjects.CreateTag,
	) (*entities.Tag, error)

	// 更新标签
	Update(
		ctx context.Context,
		userId int64,
		tagId int64,
		updateTagValueObject *valueobjects.UpdateTag,
	) error

	// 删除标签
	Delete(ctx context.Context, userId int64, tagId int64) error

	// 获取标签列表
	List(ctx context.Context, userId int64) ([]*entities.Tag, error)

	// 批量更新标签
	BatchUpdate(
		ctx context.Context,
		userId int64,
		batchUpdateTags []*valueobjects.BatchUpdateTag,
	) ([]*entities.Tag, error)

	// 获取标签偏好
	GetPreference(ctx context.Context, userId, tagId int64) (*entities.TagPreference, error)

	// 更新标签偏好
	UpdatePreference(
		ctx context.Context,
		userId int64,
		tagId int64,
		saveTagPreferenceValueObject *valueobjects.SaveTagPreference,
	) error
}

// 标签领域服务实现
type TagDomainImpl struct {
	tagRepo        repositories.TagRepository
	preferenceRepo repositories.TagPreference
}
