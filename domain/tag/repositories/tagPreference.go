package repositories

import (
	"context"
	"naotodoserver/domain/tag/vo"
)

type TagPreference interface {
	Get(ctx context.Context, whereVO *vo.TagPreference) (*vo.TagPreference, error)
	Save(ctx context.Context, preferenceVO *vo.TagPreference) (*vo.TagPreference, error)
}
