package service

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/repositories"
)

type CommentDomain interface {
	GetById(ctx context.Context, userId, commentId int64) (*entities.Comment, error)
	Create(
		ctx context.Context,
		userId int64,
		createEntity *entities.Comment,
	) (*entities.Comment, error)
	Update(
		ctx context.Context,
		userId int64,
		commentId int64,
		updateEntity *entities.Comment,
	) error
	Delete(ctx context.Context, userId, commentId int64) error
	List(ctx context.Context, userId, taskId int64) ([]*entities.Comment, error)
}

type CommentDomainImpl struct {
	CommentRepo repositories.Comment
}
