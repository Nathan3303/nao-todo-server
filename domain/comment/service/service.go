package service

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/repositories"
	"naotodoserver/domain/comment/valueobjects"
)

// CommentDomain 评论域服务接口
type CommentDomain interface {
	// GetById 获取评论详情
	// @param ctx 用户 ID
	// @param userId 评论 ID
	// @return *entities.Comment 评论实体
	// @return error 错误信息
	GetById(ctx context.Context, userId, commentId int64) (*entities.Comment, error)

	// Create 创建评论
	// @param ctx 用户 ID
	// @param userId 用户 ID
	// @param createCommentValueObject 评论值对象
	// @return *entities.Comment 评论实体
	// @return error 错误信息
	Create(
		ctx context.Context,
		userId int64,
		createCommentValueObject *valueobjects.CreateComment,
	) (*entities.Comment, error)

	// Update 更新评论
	// @param ctx 用户 ID
	// @param userId 评论 ID
	// @param updateCommentValueObject 更新评论值对象
	// @return error 错误信息
	Update(
		ctx context.Context,
		userId int64,
		commentId int64,
		updateCommentValueObject *valueobjects.UpdateComment,
	) error

	// Delete 删除评论
	// @param ctx 用户 ID
	// @param userId 评论 ID
	// @return error 错误信息
	Delete(ctx context.Context, userId, commentId int64) error

	// List 获取评论列表
	// @param ctx 用户 ID
	// @param userId 任务 ID
	// @return []*entities.Comment 评论实体列表
	// @return error 错误信息
	List(ctx context.Context, userId, taskId int64) ([]*entities.Comment, error)
}

// CommentDomainImpl 评论域服务实现
type CommentDomainImpl struct {
	CommentRepo repositories.Comment
}
