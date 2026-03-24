package service

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/vo"
)

type TagDomain interface {
	GetById(ctx context.Context, userId int64, tagId int64) (*entities.Tag, error)
	Create(ctx context.Context, userId int64, createEntity *entities.Tag) (*entities.Tag, error)
	Update(
		ctx context.Context,
		userId int64,
		tagId int64,
		updateEntity *entities.Tag,
	) error
	Delete(ctx context.Context, userId int64, tagId int64) error
	List(ctx context.Context, userId int64) ([]*entities.Tag, error)
	GetPreference(ctx context.Context, userId, tagId int64) (*vo.TagPreference, error)
	UpdatePreference(
		ctx context.Context,
		userId int64,
		tagId int64,
		preference *vo.TagPreference,
	) error
}

type TagDomainImpl struct {
	tagRepo        repositories.TagRepository
	preferenceRepo repositories.TagPreference
}
