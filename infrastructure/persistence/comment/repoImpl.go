package comment

import (
	"context"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/repositories"
	"naotodoserver/domain/comment/vo"
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
		Preload("CommentUser").
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
	createEntity *entities.Comment,
) (*entities.Comment, error) {
	// 1. 生成评论用户信息
	commentUserVO, err := commentRepo.MakeCommentUser(ctx, createEntity.UserId)
	if err != nil {
		return nil, err
	}
	// 2. 转换为模型
	createEntity.CommentUser = commentUserVO
	commentModel := CommentEntity2Model(createEntity)
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
func (commentRepo *CommentRepoImpl) Update(
	ctx context.Context,
	whereEntity *entities.Comment,
	updateEntity *entities.Comment,
) error {
	commentModel := CommentEntity2Model(updateEntity)
	whereCond := CommentEntity2Model(whereEntity)
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where(whereCond).
		Updates(commentModel).Error
	if err != nil {
		return err
	}
	return nil
}

// Delete 删除评论
func (commentRepo *CommentRepoImpl) Delete(
	ctx context.Context,
	whereEntity *entities.Comment,
) error {
	commentModel := CommentEntity2Model(whereEntity)
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Where(commentModel).
		Delete(&models.Comment{}).Error
	if err != nil {
		return err
	}
	return nil
}

// Get 获取评论列表
func (commentRepo *CommentRepoImpl) Get(
	ctx context.Context,
	whereEntity *entities.Comment,
) ([]*entities.Comment, error) {
	var commentList []*models.Comment
	err := commentRepo.db.WithContext(ctx).Model(&models.Comment{}).
		Preload("CommentUser").
		Where(CommentEntity2Model(whereEntity)).
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
) (*vo.CommentUser, error) {
	// 1. 查询用户信息
	var user models.User
	tx := commentRepo.db.Model(&models.User{}).Where("id = ?", userId).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 2. 返回评论用户
	return &vo.CommentUser{Avatar: user.Avatar, Nickname: user.Nickname}, nil

}
