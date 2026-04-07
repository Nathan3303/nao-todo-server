package service

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
)

// 创建任务清单领域服务实现
func NewProjectDomain(
	repo repositories.Project,
	preferenceRepo repositories.ProjectPreference,
) ProjectDomain {
	return &ProjectDomainImpl{
		repo:           repo,
		preferenceRepo: preferenceRepo,
	}
}

// 创建任务清单
func (p *ProjectDomainImpl) Create(
	ctx context.Context,
	createProjectValueObject *valueobjects.CreateProject,
) (*entities.Project, error) {
	return p.repo.Create(ctx, createProjectValueObject)
}

// 根据用户ID和任务清单ID获取任务清单
func (p *ProjectDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.Project, error) {
	return p.repo.GetById(ctx, userId, projectId)
}

// 更新任务清单
func (p *ProjectDomainImpl) Update(
	ctx context.Context,
	userId int64,
	projectId int64,
	updateProjectValueObject *valueobjects.UpdateProject,
) error {
	return p.repo.Update(ctx, userId, projectId, updateProjectValueObject)
}

// 删除任务清单
func (p *ProjectDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Delete(ctx, userId, projectId)
}

// 恢复任务清单
func (p *ProjectDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Restore(ctx, userId, projectId)
}

// 归档任务清单
func (p *ProjectDomainImpl) Archive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Archive(ctx, userId, projectId)
}

// 取消归档任务清单
func (p *ProjectDomainImpl) Unarchive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Unarchive(ctx, userId, projectId)
}

// 根据用户ID获取任务清单列表
func (p *ProjectDomainImpl) GetByUserId(
	ctx context.Context,
	userId int64,
) ([]*entities.Project, error) {
	return p.repo.GetByUserId(ctx, userId)
}

// 根据用户ID和任务清单ID获取任务清单偏好
func (p *ProjectDomainImpl) GetPreference(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.ProjectPreference, error) {
	return p.preferenceRepo.Get(ctx, userId, projectId)
}

// 保存任务清单偏好
func (p *ProjectDomainImpl) SavePreference(
	ctx context.Context,
	userId int64,
	projectId int64,
	saveProjectPreference *valueobjects.SaveProjectPreference,
) error {
	return p.preferenceRepo.Save(ctx, userId, projectId, saveProjectPreference)
}
