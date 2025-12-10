package comment

import (
	"context"
	"naotodoserver/domain/comment/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type CommentApp interface {
	GetComment(ctx context.Context, commentId string) (*types.CommentRes, error)
	CreateComment(ctx context.Context, req *types.CreateCommentReq) (*types.CommentRes, error)
	UpdateComment(
		ctx context.Context,
		commentId string,
		req *types.UpdateCommentReq,
	) (*types.UpdateCommentRes, error)
	DeleteComment(ctx context.Context, commentId string) (*types.DeleteCommentRes, error)
	ListComment(ctx context.Context, taskId string) (types.ListCommentRes, error)
}

type CommentAppImpl struct {
	CommentDomain service.CommentDomain
}

var (
	App  CommentApp
	once sync.Once
)
