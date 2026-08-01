package tag

import (
	"context"
	"naotodoserver/application/tag/dto"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/service"
)

// TagApp 标签应用服务接口
type TagApp interface {
	// 获取标签信息
	GetTag(ctx context.Context, tagId string) (*dto.GetTagRes, error)

	// 创建标签
	CreateTag(
		ctx context.Context,
		createTagReq *dto.CreateTagReq,
	) (*dto.CreateTagRes, error)

	// 更新标签
	UpdateTag(
		ctx context.Context,
		tagId string,
		updateTagReq *dto.UpdateTagReq,
	) error

	// 删除标签
	DeleteTag(ctx context.Context, tagId string) error

	// 获取标签列表
	ListTag(ctx context.Context) ([]*dto.GetTagRes, error)

	// 根据标签ID列表获取标签列表
	ListTagByIds(ctx context.Context, tagIds []string) ([]*dto.GetTagRes, error)

	// 批量更新标签
	BatchUpdateTags(
		ctx context.Context,
		req *dto.BatchUpdateTagReq,
	) (*dto.BatchUpdateTagRes, error)

	// 获取标签偏好设置
	GetTagPreference(
		ctx context.Context,
		tagId string,
	) (*dto.GetTagPreferenceRes, error)

	// 更新标签偏好设置
	UpdateTagPreference(
		ctx context.Context,
		tagId string,
		updateTagPreferenceReq *dto.UpdateTagPreferenceReq,
	) error
}

// TagAppImpl 标签应用服务实现
type TagAppImpl struct {
	tagDomain      service.TagDomain
	tagRepo        repositories.TagRepository
	preferenceRepo repositories.TagPreference
}
