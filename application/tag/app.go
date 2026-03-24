package tag

import (
	"context"
	"naotodoserver/domain/tag/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type TagApp interface {
	GetTag(ctx context.Context, tagId string) (*types.GetTagRes, error)
	CreateTag(ctx context.Context, req *types.CreateTagReq) (*types.CreateTagRes, error)
	UpdateTag(ctx context.Context, req *types.UpdateTagReq) (*types.UpdateTagRes, error)
	DeleteTag(ctx context.Context, tagId string) (*types.DeleteTagRes, error)
	ListTag(ctx context.Context) (types.ListTagRes, error)
	GetTagPreference(
		ctx context.Context,
		tagId string,
	) (*types.TagPreferenceRes, error)
	UpdateTagPreference(
		ctx context.Context,
		tagId string,
		req *types.UpdateTagPreferenceReq,
	) (*types.UpdateTagPreferenceRes, error)
}

type TagAppImpl struct {
	tagDomain service.TagDomain
}

var (
	App  TagAppImpl
	once sync.Once
)
