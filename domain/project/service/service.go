package service

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
)

// ProjectDomain 任务清单领域服务接口
type ProjectDomain interface {
	// 创建任务清单
	Create(
		ctx context.Context,
		createProjectValueObject *valueobjects.CreateProject,
	) (*entities.Project, error)

	// 删除任务清单
	Delete(ctx context.Context, userId int64, projectId int64) error

	// 恢复任务清单
	Restore(ctx context.Context, userId int64, projectId int64) error
}

// ProjectDomainImpl 任务清单领域服务实现
type ProjectDomainImpl struct {
	repo           repositories.Project
	preferenceRepo repositories.ProjectPreference
}
