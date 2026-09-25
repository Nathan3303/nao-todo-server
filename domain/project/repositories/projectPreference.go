package repositories

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
)

// ProjectPreference 任务清单偏好仓库接口
type ProjectPreference interface {
	// Get 获取任务清单偏好
	Get(ctx context.Context, userId int64, projectId int64) (*entities.ProjectPreference, error)

	// Save 保存任务清单偏好
	Save(
		ctx context.Context,
		userId int64,
		projectId int64,
		saveProjectPreference *valueobjects.SaveProjectPreference,
	) error

	// Delete 删除任务清单偏好
	Delete(ctx context.Context, userId int64, projectId int64) error

	// Restore 恢复任务清单偏好
	Restore(ctx context.Context, userId int64, projectId int64) error
}
