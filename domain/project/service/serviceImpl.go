package service

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/vo"
)

func NewProjectDomain(
	repo repositories.Project,
	preferenceRepo repositories.ProjectPreference,
) ProjectDomain {
	return &ProjectDomainImpl{
		repo:           repo,
		preferenceRepo: preferenceRepo,
	}
}

/*
 * Create project
 */
func (p *ProjectDomainImpl) Create(
	ctx context.Context,
	projectEntity *entities.Project,
) (*entities.Project, error) {
	// projectPreference := vo.MakeDefaultProjectPreference(projectEntity.UserId)
	// projectEntity.Preference = projectPreference
	return p.repo.Create(ctx, projectEntity)
}

/*
 * Get project by userId and projectId
 */
func (p *ProjectDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.Project, error) {
	return p.repo.GetById(ctx, userId, projectId)
}

/*
 * Update project
 */
func (p *ProjectDomainImpl) Update(
	ctx context.Context,
	userId int64,
	projectId int64,
	updateEntity *entities.Project,
) error {
	return p.repo.Update(
		ctx,
		&entities.Project{UserId: userId, Id: projectId},
		updateEntity,
	)
}

/*
 * Delete project
 */
func (p *ProjectDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Delete(ctx, &entities.Project{UserId: userId, Id: projectId})
}

/*
 * Restore project
 */
func (p *ProjectDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Restore(ctx, &entities.Project{UserId: userId, Id: projectId})
}

/*
 * Archive project
 */
func (p *ProjectDomainImpl) Archive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Archive(ctx, &entities.Project{UserId: userId, Id: projectId})
}

/*
 * Unarchive project
 */
func (p *ProjectDomainImpl) Unarchive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	return p.repo.Unarchive(ctx, &entities.Project{UserId: userId, Id: projectId})
}

/*
 * Get projects by userId
 */
func (p *ProjectDomainImpl) GetByUserId(
	ctx context.Context,
	userId int64,
) ([]*entities.Project, error) {
	return p.repo.GetByUserId(ctx, userId)
}

/*
 * Get project preference by userId and projectId
 */
func (p *ProjectDomainImpl) GetPreference(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*vo.ProjectPreference, error) {
	var preferenceVO vo.ProjectPreference
	preferenceVO.UserId = userId
	preferenceVO.ProjectId = projectId
	return p.preferenceRepo.Get(ctx, preferenceVO)
}

/*
 * Save project preference
 */
func (p *ProjectDomainImpl) SavePreference(
	ctx context.Context,
	userId int64,
	projectId int64,
	preferenceVO *vo.ProjectPreference,
) error {
	preferenceVO.UserId = userId
	preferenceVO.ProjectId = projectId
	_, err := p.preferenceRepo.Save(ctx, preferenceVO)
	return err
}
