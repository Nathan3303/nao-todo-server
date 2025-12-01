package auth

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/infrastructure/persistence/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepoImpl struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repositories.User {
	return &userRepoImpl{db: db}
}

/**
 * Create User
 */
func (ur *userRepoImpl) Create(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	// 1. 实体转换
	userModel := UserEntity2Model(userEntity)
	// 2. 创建用户
	tx := ur.db.WithContext(ctx).Model(&models.User{}).Create(userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 模型转换并返回
	return UserModel2Entity(userModel), nil
}

/**
 * Find User By Email
 */
func (ur *userRepoImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	// 1. 创建结果模型
	user := &models.User{}
	// 2. 创建查找模型
	userModel := &models.User{Email: email}
	// 3. 执行查找
	tx := ur.db.WithContext(ctx).Model(&models.User{}).Where(userModel).First(&user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 模型转换并返回
	return UserModel2Entity(user), nil
}

/**
 * Compare Password
 */
func (u *userRepoImpl) PasswordCompare(password []byte, encryptedPassword []byte) bool {
	return bcrypt.CompareHashAndPassword(encryptedPassword, password) == nil
}
