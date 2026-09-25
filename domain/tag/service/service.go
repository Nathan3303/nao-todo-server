package service

import (
	"context"

	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

// TagDomain 标签领域服务接口
type TagDomain interface {
	// 创建标签
	// 返回 UpsertResult：Created=true 表示本次为新建（含墓碑复活）；Outcome 供同步回执使用
	Create(
		ctx context.Context,
		userId int64,
		createTagValueObject *valueobjects.CreateTag,
	) (*entities.Tag, domaintypes.UpsertResult, error)

	// 删除标签
	Delete(ctx context.Context, userId int64, tagId int64) error
}

// TagDomainImpl 标签领域服务实现
type TagDomainImpl struct {
	tagRepo        repositories.TagRepository
	preferenceRepo repositories.TagPreference
}
