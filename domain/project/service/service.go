package service

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/vo"
)

type ProjectDomain interface {
	Create(ctx context.Context, projectEntity *entities.Project) (*entities.Project, error)
	GetById(ctx context.Context, userId int64, projectId int64) (*entities.Project, error)
	Update(ctx context.Context, userId int64, projectId int64, updateEntity *entities.Project) error
	Delete(ctx context.Context, userId int64, projectId int64) error
	Restore(ctx context.Context, userId int64, projectId int64) error
	Archive(ctx context.Context, userId int64, projectId int64) error
	Unarchive(ctx context.Context, userId int64, projectId int64) error
	GetByUserId(ctx context.Context, userId int64) ([]*entities.Project, error)
	GetPreference(ctx context.Context, userId, projectId int64) (*vo.ProjectPreference, error)
}

type ProjectDomainImpl struct {
	repo           repositories.Project
	preferenceRepo repositories.ProjectPreference
}
