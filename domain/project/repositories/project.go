package repositories

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/domain/types"
)

// Project 任务清单仓库接口
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

	// 更新任务清单状态（归档/停用时间字段）
	UpdateState(
		ctx context.Context,
		userId int64,
		projectId int64,
		archivedAt types.NullableTime,
		deactivedAt types.NullableTime,
	) error

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

	// 删除已注销的任务清单
	DeleteDeactivatedProjects(ctx context.Context, dayOffset int8) (int64, error)
}
