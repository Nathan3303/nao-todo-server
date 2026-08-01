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
	// 1. 加载任务清单实体
	projectEntity, err := p.repo.GetById(ctx, userId, projectId)
	if err != nil {
		return err
	}
	// 2. 应用实体删除方法（软删除，设置停用时间）
	projectEntity.Delete()
	// 3. 持久化实体状态
	err = p.repo.UpdateState(
		ctx,
		userId,
		projectId,
		projectEntity.ArchivedAt,
		projectEntity.DeactivedAt,
	)
	if err != nil {
		return err
	}
	// 4. 删除任务清单偏好
	return p.preferenceRepo.Delete(ctx, userId, projectId)
}

// Restore 恢复任务清单
func (p *ProjectDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 加载任务清单实体
	projectEntity, err := p.repo.GetById(ctx, userId, projectId)
	if err != nil {
		return err
	}
	// 2. 应用实体恢复方法（清空停用时间）
	projectEntity.Restore()
	// 3. 持久化实体状态
	err = p.repo.UpdateState(
		ctx,
		userId,
		projectId,
		projectEntity.ArchivedAt,
		projectEntity.DeactivedAt,
	)
	if err != nil {
		return err
	}
	// 4. 恢复任务清单偏好
	return p.preferenceRepo.Restore(ctx, userId, projectId)
}

// Archive 归档任务清单
func (p *ProjectDomainImpl) Archive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 加载任务清单实体
	projectEntity, err := p.repo.GetById(ctx, userId, projectId)
	if err != nil {
		return err
	}
	// 2. 应用实体归档方法（幂等重设归档时间）
	projectEntity.Archive()
	// 3. 持久化实体状态
	return p.repo.UpdateState(
		ctx,
		userId,
		projectId,
		projectEntity.ArchivedAt,
		projectEntity.DeactivedAt,
	)
}

// Unarchive 取消归档任务清单
func (p *ProjectDomainImpl) Unarchive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 加载任务清单实体
	projectEntity, err := p.repo.GetById(ctx, userId, projectId)
	if err != nil {
		return err
	}
	// 2. 应用实体取消归档方法（清空归档时间）
	projectEntity.Unarchive()
	// 3. 持久化实体状态
	return p.repo.UpdateState(
		ctx,
		userId,
		projectId,
		projectEntity.ArchivedAt,
		projectEntity.DeactivedAt,
	)
}
