package auth

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/infrastructure/persistence/models"
	"time"

	"gorm.io/gorm"
)

type sessionRepoImpl struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) repositories.Session {
	return &sessionRepoImpl{
		db: db,
	}
}

/**
 * Create Session
 */
func (sr *sessionRepoImpl) Create(ctx context.Context, sessionEntity *entities.Session) error {
	// 1. 查找现存记录
	session := sr.FindByUserIdAndToken(ctx, sessionEntity.UserId, sessionEntity.Token)
	// 2. 执行结果存在逻辑
	if session.IsValid() {
		// 更新记录

		tx := sr.db.WithContext(ctx).Model(&models.Session{}).Where(&models.Session{
			UserId: userEntity.Id,
			Token:  session.Token,
		}).UpdateColumns(&UserSession{
			Token:     token,
			ExpiredAt: time.Now().Add(time.Hour * 48),
		})
		if tx.Error != nil {
			return tx.Error
		}
		return nil
	}
	// 3. 执行结果不存在逻辑
	tx := r.db.WithContext(ctx).Model(&UserSession{}).Create(&UserSession{
		UserId:     userEntity.Id,
		Token:      token,
		ExpiredAt:  time.Now().Add(time.Hour * 48),
		DeviceType: "webapp",
	})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Delete implements repositories.Session.
func (sr *sessionRepoImpl) Delete(ctx context.Context, userId int64, token string) error {
	// 1. 创建删除模型
	deleteCond := &UserSession{Token: token}
	// 2. 执行删除
	tx := r.db.WithContext(ctx).Model(&UserSession{}).Where(&deleteCond).Delete(&UserSession{})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 返回结果
	return nil
}

// FindByUserIdAndToken implements repositories.Session.
func (sr *sessionRepoImpl) FindByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) *entities.Session {
	// 1. 创建结果模型
	session := &UserSession{}
	// 2. 创建查找模型
	findCond := &UserSession{Token: token}
	// 3. 执行查找
	r.db.WithContext(ctx).Model(&UserSession{}).Where(findCond).First(&session)
	// 4. 模型转换并返回
	return UserSessionModelToEntity(session)
}

// UpdateToken implements repositories.Session.
func (sr *sessionRepoImpl) UpdateToken(ctx context.Context, sessionEntity *entities.Session) error {
	// 1. 创建更新模型
	updateCond := &UserSession{
		UserId: userSession.UserId,
		Token:  userSession.Token,
	}
	// 2. 执行更新
	tx := r.db.WithContext(ctx).Model(&UserSession{}).Where(updateCond).Updates(&UserSession{
		Token: updateSession.Token,
	})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 返回结果
	return nil
}
