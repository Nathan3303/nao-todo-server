package repositories

import (
	"context"
	"naotodoserver/domain/project/vo"
)

type ProjectPreference interface {
	Get(ctx context.Context, preferenceVO vo.ProjectPreference) (*vo.ProjectPreference, error)
	Save(ctx context.Context, preferenceVO *vo.ProjectPreference) (*vo.ProjectPreference, error)
}
