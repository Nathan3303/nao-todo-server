package service

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/repositories"
)

func NewCommentDomain(commentRepo repositories.Comment) CommentDomain {
	return &CommentDomainImpl{
		CommentRepo: commentRepo,
	}
}

// GetById 获取评论详情
func (commentDomain *CommentDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	commentId int64,
) (*entities.Comment, error) {
	return commentDomain.CommentRepo.GetById(ctx, userId, commentId)
}

// Create 创建评论
func (commentDomain *CommentDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEntity *entities.Comment,
) (*entities.Comment, error) {
	createEntity.UserId = userId
	return commentDomain.CommentRepo.Create(ctx, createEntity)
}

// Update 更新评论
func (commentDomain *CommentDomainImpl) Update(
	ctx context.Context,
	userId int64,
	commentId int64,
	updateEntity *entities.Comment,
) error {
	whereEntity := &entities.Comment{UserId: userId, Id: commentId}
	return commentDomain.CommentRepo.Update(ctx, whereEntity, updateEntity)
}

// Delete 删除评论
func (commentDomain *CommentDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	commentId int64,
) error {
	whereEntity := &entities.Comment{UserId: userId, Id: commentId}
	return commentDomain.CommentRepo.Delete(ctx, whereEntity)
}

// List 获取评论列表
func (commentDomain *CommentDomainImpl) List(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Comment, error) {
	whereEntity := &entities.Comment{UserId: userId, TaskId: taskId}
	return commentDomain.CommentRepo.Get(ctx, whereEntity)
}
