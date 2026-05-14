package repositories

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
)

// 任务清单仓库接口
type Project interface {
	// 创建任务清单
	Create(
		ctx context.Context,
		createProjectValueObject *valueobjects.CreateProject,
	) (*entities.Project, error)

	// 根据用户ID和任务清单ID获取任务清单
	GetById(ctx context.Context, userId int64, projectId int64) (*entities.Project, error)

	// 更新任务清单
	Update(
		ctx context.Context,
		userId int64,
		projectId int64,
		updateProjectValueObject *valueobjects.UpdateProject,
	) error

	// 删除任务清单
	Delete(ctx context.Context, userId int64, projectId int64) error

	// 恢复任务清单
	Restore(ctx context.Context, userId int64, projectId int64) error

	// 归档任务清单
	Archive(ctx context.Context, userId int64, projectId int64) error

	// 取消归档任务清单
	Unarchive(ctx context.Context, userId int64, projectId int64) error

	// 根据用户ID获取任务清单列表
	GetByUserId(ctx context.Context, userId int64) ([]*entities.Project, error)

	// 获取最大排序 ID
	GetMaxSortId(ctx context.Context, userId int64) uint16

	// 批量更新任务清单
	BatchUpdate(
		ctx context.Context,
		userId int64,
		batchUpdateProjects []*valueobjects.BatchUpdateProject,
	) ([]*entities.Project, error)
}
