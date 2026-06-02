package repositories

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/valueobjects"
)

// Comment 评论仓库接口
type Comment interface {
	// GetById 获取评论详情
	// @param ctx 用户 ID
	// @param userId 评论 ID
	// @return *entities.Comment 评论实体
	// @return error 错误信息
	GetById(ctx context.Context, userId, commentId int64) (*entities.Comment, error)

	// Create 创建评论
	// @param ctx 用户 ID
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
	// @param commentId 评论 ID
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
	// @param commentId 评论 ID
	// @return error 错误信息
	Delete(ctx context.Context, userId, commentId int64) error

	// Get 获取评论列表
	Get(ctx context.Context, userId, taskId int64) ([]*entities.Comment, error)

	// SyncUserProfile 同步用户资料到该用户所有历史评论
	SyncUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}
