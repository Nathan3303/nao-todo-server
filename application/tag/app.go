package tag

import (
	"context"
	"naotodoserver/domain/tag/service"
	"naotodoserver/interfaces/types"
)

// 标签应用服务接口
type TagApp interface {
	// 获取标签信息
	GetTag(ctx context.Context, tagId string) (*types.GetTagRes, error)

	// 创建标签
	CreateTag(
		ctx context.Context,
		createTagReq *types.CreateTagReq,
	) (*types.CreateTagRes, error)

	// 更新标签
	UpdateTag(
		ctx context.Context,
		tagId string,
		updateTagReq *types.UpdateTagReq,
	) error

	// 删除标签
	DeleteTag(ctx context.Context, tagId string) error

	// 获取标签列表
	ListTag(ctx context.Context) ([]*types.GetTagRes, error)

	// 批量更新标签
	BatchUpdateTags(ctx context.Context, req *types.BatchUpdateTagReq) (*types.BatchUpdateTagRes, error)

	// 获取标签偏好设置
	GetTagPreference(
		ctx context.Context,
		tagId string,
	) (*types.GetTagPreferenceRes, error)

	// 更新标签偏好设置
	UpdateTagPreference(
		ctx context.Context,
		tagId string,
		updateTagPreferenceReq *types.UpdateTagPreferenceReq,
	) error
}

// 标签应用服务实现
type TagAppImpl struct {
	tagDomain service.TagDomain
}

// 标签应用服务单例
var App TagApp
