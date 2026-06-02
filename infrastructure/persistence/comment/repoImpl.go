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
func (commentRepo *CommentRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	commentId int64,
) (*entities.Comment, error) {
	var comment models.Comment
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where("user_id = ? AND id = ?", userId, commentId).
		First(&comment).Error
	if err != nil {
		return nil, err
	}
	return CommentModel2Entity(&comment), nil
}

// Create 创建评论
func (commentRepo *CommentRepoImpl) Create(
	ctx context.Context,
	userId int64,
	createCommentValueObject *valueobjects.CreateComment,
) (*entities.Comment, error) {
	// 1. 获取用户昵称和头像
	var user models.User
	tx := commentRepo.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userId).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 2. 转换为模型并填入用户信息
	commentModel := CreateCommentValueObjectToModel(createCommentValueObject)
	commentModel.Nickname = user.Nickname
	commentModel.Avatar = user.Avatar
	// 3. 创建评论
	err := commentRepo.db.WithContext(ctx).Create(commentModel).Error
	if err != nil {
		return nil, err
	}
	// 4. 返回评论实体
	return CommentModel2Entity(commentModel), nil
}

// Update 更新评论
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
func (commentRepo *CommentRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	commentId int64,
) error {
	var whereCond models.Comment
	whereCond.UserId = userId
	whereCond.ID = commentId
	return commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where(&whereCond).
		Delete(&models.Comment{}).Error
}

// Get 获取评论列表
func (commentRepo *CommentRepoImpl) Get(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.Comment, error) {
	var commentList []*models.Comment
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where("user_id = ? AND task_id = ?", userId, taskId).
		Find(&commentList).Error
	if err != nil {
		return nil, err
	}
	return CommentModels2Entities(commentList), nil
}

// SyncUserProfile 同步用户资料到历史评论
func (commentRepo *CommentRepoImpl) SyncUserProfile(
	ctx context.Context,
	userId int64,
	nickname string,
	avatar string,
) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return nil
	}
	return commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where("user_id = ?", userId).
		Updates(updates).Error
}
