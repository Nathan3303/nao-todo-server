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
