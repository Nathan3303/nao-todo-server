package repositories

import (
	"context"
	"naotodoserver/domain/project/entities"
)

type Project interface {
	Create(ctx context.Context, project *entities.Project) (*entities.Project, error)
	GetById(ctx context.Context, userId int64, projectId int64) (*entities.Project, error)
	Update(ctx context.Context, whereEntity *entities.Project, updateEntity *entities.Project) error
	Delete(ctx context.Context, whereEntity *entities.Project) error
	Restore(ctx context.Context, whereEntity *entities.Project) error
	Archive(ctx context.Context, whereEntity *entities.Project) error
	Unarchive(ctx context.Context, whereEntity *entities.Project) error
	GetByUserId(ctx context.Context, userId int64) ([]*entities.Project, error)
}
