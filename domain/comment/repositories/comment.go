package repositories

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/vo"
)

type Comment interface {
	GetById(ctx context.Context, userId, commentId int64) (*entities.Comment, error)
	Create(ctx context.Context, createEntity *entities.Comment) (*entities.Comment, error)
	Update(ctx context.Context, whereEntity, updateEntity *entities.Comment) error
	Delete(ctx context.Context, whereEntity *entities.Comment) error
	Get(ctx context.Context, whereEntity *entities.Comment) ([]*entities.Comment, error)
	MakeCommentUser(ctx context.Context, userId int64) (*vo.CommentUser, error)
}
