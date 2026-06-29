package service

import (
	"context"
	"fmt"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
)

// NewProjectDomain 创建任务清单领域服务实现
func NewProjectDomain(
	repo repositories.Project,
	preferenceRepo repositories.ProjectPreference,
) ProjectDomain {
	return &ProjectDomainImpl{
		repo:           repo,
		preferenceRepo: preferenceRepo,
	}
}

// Create 创建任务清单
func (p *ProjectDomainImpl) Create(
	ctx context.Context,
	createProjectValueObject *valueobjects.CreateProject,
) (*entities.Project, error) {
	// 设置排序 ID
	createProjectValueObject.SortId = p.repo.GetMaxSortId(ctx, createProjectValueObject.UserId) + 1
	// 创建任务清单
	projectEntity, err := p.repo.Create(ctx, createProjectValueObject)
	if err != nil {
		return nil, err
	}
	// 创建任务清单基础偏好值对象
	projectPreferenceValueObject, err := valueobjects.NewSaveProjectPreference(
		"table",
		fmt.Sprintf("{\"projectId\": \"%d\"}", projectEntity.Id),
		"{}",
	)
	if err != nil {
		return nil, err
	}
	// 保存任务清单基础偏好值对象
	err = p.preferenceRepo.Save(
		ctx,
		createProjectValueObject.UserId,
		projectEntity.Id,
		projectPreferenceValueObject,
	)
	if err != nil {
		return nil, err
	}
	// 返回任务清单实体
	return projectEntity, nil
}

// GetById 根据用户ID和任务清单ID获取任务清单
func (p *ProjectDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.Project, error) {
	return p.repo.GetById(ctx, userId, projectId)
}

// Update 更新任务清单
func (p *ProjectDomainImpl) Update(
	ctx context.Context,
	userId int64,
	projectId int64,
	updateProjectValueObject *valueobjects.UpdateProject,
) error {
	return p.repo.Update(ctx, userId, projectId, updateProjectValueObject)
}

// Delete 删除任务清单
func (p *ProjectDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 删除任务清单
	err := p.repo.Delete(ctx, userId, projectId)
	if err != nil {
		return err
	}
	// 删除任务清单偏好
	return p.preferenceRepo.Delete(ctx, userId, projectId)
}

// Restore 恢复任务清单
func (p *ProjectDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 恢复任务清单
	err := p.repo.Restore(ctx, userId, projectId)
	if err != nil {
		return err
	}
	// 恢复任务清单偏好
	return p.preferenceRepo.Restore(ctx, userId, projectId)
}

// Archive 归档任务清单
func (p *ProjectDomainImpl) Archive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Archive(ctx, userId, projectId)
}

// Unarchive 取消归档任务清单
func (p *ProjectDomainImpl) Unarchive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Unarchive(ctx, userId, projectId)
}

// GetByUserId 根据用户ID获取任务清单列表
func (p *ProjectDomainImpl) GetByUserId(
	ctx context.Context,
	userId int64,
) ([]*entities.Project, error) {
	return p.repo.GetByUserId(ctx, userId)
}

// BatchUpdate 批量更新任务清单
func (p *ProjectDomainImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateProjects []*valueobjects.BatchUpdateProject,
) ([]*entities.Project, error) {
	return p.repo.BatchUpdate(ctx, userId, batchUpdateProjects)
}

// GetPreference 根据用户ID和任务清单ID获取任务清单偏好
func (p *ProjectDomainImpl) GetPreference(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.ProjectPreference, error) {
	return p.preferenceRepo.Get(ctx, userId, projectId)
}

// SavePreference 保存任务清单偏好
func (p *ProjectDomainImpl) SavePreference(
	ctx context.Context,
	userId int64,
	projectId int64,
	saveProjectPreference *valueobjects.SaveProjectPreference,
) error {
	return p.preferenceRepo.Save(ctx, userId, projectId, saveProjectPreference)
}

// DeleteDeactivatedProjects 删除已注销的任务清单（供定时任务调用）
func (p *ProjectDomainImpl) DeleteDeactivatedProjects(
	ctx context.Context,
	dayOffset int8,
) (int64, error) {
	return p.repo.DeleteDeactivatedProjects(ctx, dayOffset)
}
