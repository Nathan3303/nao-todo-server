package repositories

import (
	"context"
	"naotodoserver/domain/tag/entities"
)

type TagRepository interface {
	GetById(ctx context.Context, userId int64, tagId int64) (*entities.Tag, error)
	Create(ctx context.Context, createEntity *entities.Tag) (*entities.Tag, error)
	Update(ctx context.Context, whereEntity *entities.Tag, updateEntity *entities.Tag) error
	Delete(ctx context.Context, whereEntity *entities.Tag) error
	Get(ctx context.Context, whereEntity *entities.Tag) ([]*entities.Tag, error)
}
