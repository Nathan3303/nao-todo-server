package comment

import (
	"context"
	"naotodoserver/domain/comment/service"
	"naotodoserver/interfaces/types"
)

// CommentApp 评论应用接口
type CommentApp interface {
	// GetComment 获取评论详情
	// @param ctx 上下文
	// @return comment 评论详情
	// @return err 错误
	GetComment(ctx context.Context, commentId string) (*types.CommentRes, error)

	// CreateComment 创建评论
	// @param ctx 上下文
	// @return comment 评论详情
	// @return err 错误
	CreateComment(ctx context.Context, req *types.CreateCommentReq) (*types.CommentRes, error)

	// UpdateComment 更新评论
	// @param ctx 上下文
	// @param commentId 评论 ID
	// @param req 更新评论请求
	// @return err 错误
	UpdateComment(
		ctx context.Context,
		commentId string,
		req *types.UpdateCommentReq,
	) error

	// DeleteComment 删除评论
	// @param ctx 上下文
	// @param commentId 评论 ID
	// @return err 错误
	DeleteComment(ctx context.Context, commentId string) error

	// ListComment 获取评论列表
	// @param ctx 上下文
	// @param taskId 任务 ID
	// @return comments 评论列表
	// @return err 错误
	ListComment(ctx context.Context, taskId string) ([]*types.CommentRes, error)
}

// CommentAppImpl 评论应用实现
type CommentAppImpl struct {
	CommentDomain service.CommentDomain
}

// App 评论应用实例
