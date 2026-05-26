package comment

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/repositories"
	"naotodoserver/domain/comment/valueobjects"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type CommentRepoImpl struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) repositories.Comment {
	return &CommentRepoImpl{
		db: db,
	}
}

// GetById 获取评论详情
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @return 评论实体
// @return error 错误
func (commentRepo *CommentRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	commentId int64,
) (*entities.Comment, error) {
	var comment models.Comment
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Preload("CommentUser").
		Where("user_id = ? AND id = ?", userId, commentId).
		First(&comment).Error
	if err != nil {
		return nil, err
	}
	return CommentModel2Entity(&comment), nil
}

// Create 创建评论
// @param ctx 上下文
// @param userId 用户ID
// @param createCommentValueObject 创建评论值对象
// @return 评论实体
// @return error 错误
func (commentRepo *CommentRepoImpl) Create(
	ctx context.Context,
	userId int64,
	createCommentValueObject *valueobjects.CreateComment,
) (*entities.Comment, error) {
	// 1. 生成评论用户信息
	commentUserVO, err := commentRepo.MakeCommentUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 2. 转换为模型
	commentModel := CreateCommentValueObjectToModel(createCommentValueObject)
	commentModel.CommentUser = &models.CommentUser{
		CommentId: commentModel.ID,
		Avatar:    commentUserVO.Avatar,
		Nickname:  commentUserVO.Nickname,
	}
	// 3. 创建评论
	err = commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Preload("CommentUser").
		Create(commentModel).Error
	if err != nil {
		return nil, err
	}
	// 4. 返回评论实体
	return CommentModel2Entity(commentModel), nil
}

// Update 更新评论
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @param updateCommentValueObject 更新评论值对象
// @return error 错误
func (commentRepo *CommentRepoImpl) Update(
	ctx context.Context,
	userId int64,
	commentId int64,
	updateCommentValueObject *valueobjects.UpdateComment,
) error {
	var whereCond models.Comment
	whereCond.UserId = userId
	whereCond.ID = commentId
	updateCond := UpdateCommentValueObjectToMap(updateCommentValueObject)
	return commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where(&whereCond).
		Updates(updateCond).Error
}

// Delete 删除评论
// @param ctx 上下文
// @param userId 用户ID
// @param commentId 评论ID
// @return error 错误
func (commentRepo *CommentRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	commentId int64,
) error {
	var whereCond models.Comment
	whereCond.UserId = userId
	whereCond.ID = commentId
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where(&whereCond).
		Delete(&models.Comment{}).Error
	if err != nil {
		return err
	}
	return nil
}

// Get 获取评论列表
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 待办任务ID
// @return 评论实体列表
// @return error 错误
func (commentRepo *CommentRepoImpl) Get(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Comment, error) {
	var commentList []*models.Comment
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Preload("CommentUser").
		Where("user_id = ? AND task_id = ?", userId, taskId).
		Find(&commentList).Error
	if err != nil {
		return nil, err
	}
	return CommentModels2Entities(commentList), nil
}

// GetCommentUser 获取评论用户
func (commentRepo *CommentRepoImpl) MakeCommentUser(
	ctx context.Context,
	userId int64,
) (*entities.CommentUser, error) {
	// 1. 查询用户信息
	var user models.User
	tx := commentRepo.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userId).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 2. 返回评论用户
	return &entities.CommentUser{Avatar: user.Avatar, Nickname: user.Nickname}, nil

}
