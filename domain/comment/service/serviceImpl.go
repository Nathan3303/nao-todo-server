package service

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/repositories"
	"naotodoserver/domain/comment/valueobjects"
)

// NewCommentDomain 创建评论域服务实现
func NewCommentDomain(commentRepo repositories.Comment) CommentDomain {
	return &CommentDomainImpl{
		CommentRepo: commentRepo,
	}
}

// GetById 获取评论详情
// @param ctx 用户 ID
// @param userId 评论 ID
// @return *entities.Comment 评论实体
// @return error 错误信息
func (commentDomain *CommentDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	commentId int64,
) (*entities.Comment, error) {
	return commentDomain.CommentRepo.GetById(ctx, userId, commentId)
}

// Create 创建评论
// @param ctx 用户 ID
// @param userId 用户 ID
// @param createCommentValueObject 评论值对象
// @return *entities.Comment 评论实体
// @return error 错误信息
func (commentDomain *CommentDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createCommentValueObject *valueobjects.CreateComment,
) (*entities.Comment, error) {
	return commentDomain.CommentRepo.Create(ctx, userId, createCommentValueObject)
}

// Update 更新评论
// @param ctx 用户 ID
// @param userId 评论 ID
// @param updateCommentValueObject 更新评论值对象
// @return error 错误信息
func (commentDomain *CommentDomainImpl) Update(
	ctx context.Context,
	userId int64,
	commentId int64,
	updateCommentValueObject *valueobjects.UpdateComment,
) error {
	return commentDomain.CommentRepo.Update(ctx, userId, commentId, updateCommentValueObject)
}

// Delete 删除评论
// @param ctx 用户 ID
// @param userId 评论 ID
// @param commentId 评论 ID
// @return error 错误信息
func (commentDomain *CommentDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	commentId int64,
) error {
	return commentDomain.CommentRepo.Delete(ctx, userId, commentId)
}

// List 获取评论列表
// @param ctx 用户 ID
// @param userId 用户 ID
// @param taskId 待办任务 ID
// @return []*entities.Comment 评论实体列表
// @return error 错误信息
func (commentDomain *CommentDomainImpl) List(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Comment, error) {
	return commentDomain.CommentRepo.Get(ctx, userId, taskId)
}
