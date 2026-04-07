package service

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
)

// 任务清单领域服务接口
type ProjectDomain interface {
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

	// 获取任务清单偏好
	GetPreference(
		ctx context.Context,
		userId int64,
		projectId int64,
	) (*entities.ProjectPreference, error)

	// 根据用户ID和任务清单ID保存任务清单偏好
	SavePreference(
		ctx context.Context,
		userId int64,
		projectId int64,
		saveProjectPreference *valueobjects.SaveProjectPreference,
	) error
}

// 任务清单领域服务实现
type ProjectDomainImpl struct {
	repo           repositories.Project
	preferenceRepo repositories.ProjectPreference
}
