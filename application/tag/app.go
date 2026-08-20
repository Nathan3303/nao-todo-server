package tag

import (
	"context"
	"naotodoserver/application/tag/dto"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/service"
	taskService "naotodoserver/domain/task/service"
	"naotodoserver/domain/types"
)

// TagApp 标签应用服务接口
type TagApp interface {
	// 获取标签信息
	GetTag(ctx context.Context, userId int64, tagId string) (*dto.GetTagRes, error)

	// 创建标签
	CreateTag(
		ctx context.Context,
		userId int64,
		createTagReq *dto.CreateTagReq,
	) (*dto.CreateTagRes, error)

	// 更新标签
	UpdateTag(
		ctx context.Context,
		userId int64,
		tagId string,
		updateTagReq *dto.UpdateTagReq,
	) error

	// 删除标签
	DeleteTag(ctx context.Context, userId int64, tagId string) error

	// 获取标签列表
	ListTag(ctx context.Context, userId int64) ([]*dto.GetTagRes, error)
	// ListTagSync 增量同步标签列表（包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页）
	ListTagSync(
		ctx context.Context,
		userId int64,
		updatedAt string,
		cursorId string,
		limit int,
	) ([]*dto.GetTagRes, error)

	// 根据标签ID列表获取标签列表
	ListTagByIds(ctx context.Context, userId int64, tagIds []string) ([]*dto.GetTagRes, error)

	// 批量更新标签
	BatchUpdateTags(
		ctx context.Context,
		userId int64,
		req *dto.BatchUpdateTagReq,
	) (*dto.BatchUpdateTagRes, error)

	// 获取标签偏好设置
	GetTagPreference(
		ctx context.Context,
		userId int64,
		tagId string,
	) (*dto.GetTagPreferenceRes, error)

	// 更新标签偏好设置
	UpdateTagPreference(
		ctx context.Context,
		userId int64,
		tagId string,
		updateTagPreferenceReq *dto.UpdateTagPreferenceReq,
	) error
}

// TagAppImpl 标签应用服务实现
type TagAppImpl struct {
	tagDomain      service.TagDomain
	tagRepo        repositories.TagRepository
	preferenceRepo repositories.TagPreference
	txManager      types.TxManager
	taskDomain     taskService.TaskDomain
}
