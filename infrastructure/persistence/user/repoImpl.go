package user

import (
	"context"
	"errors"
	"naotodoserver/domain/user/entities"
	"naotodoserver/domain/user/repositories"
	"naotodoserver/infrastructure/persistence/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepoImpl struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) repositories.User {
	return &UserRepoImpl{db: db}
}

/**
 * Create User
 */
func (userRepo *UserRepoImpl) Create(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	// 1. 验证模型
	if !userEntity.IsValid() {
		return nil, errors.New("参数无效")
	}
	// 1. 创建模型
	userModel := UserEntity2Model(userEntity)
	// 2. 创建用户
	tx := userRepo.db.WithContext(ctx).Model(&models.User{}).Create(&userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 模型转换并返回
	return UserModel2Entity(userModel), nil
}

/**
 * Find User By Id
 */
func (userRepo *UserRepoImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	// 1. 创建结果模型
	userModel := &models.User{}
	// 2. 创建查找模型
	findCond := &models.User{}
	findCond.ID = id
	// 3. 执行查找
	tx := userRepo.db.WithContext(ctx).Model(&models.User{}).Where(findCond).First(&userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 模型转换并返回
	return UserModel2Entity(userModel), nil
}

/**
 * Update User Avatar
 */
func (userRepo *UserRepoImpl) UpdateAvatar(
	ctx context.Context,
	userId int64,
	avatarUrl string,
) error {
	// 1. 创建更新模型
	updateModel := &models.User{Avatar: avatarUrl}
	// 2. 创建查找模型
	findCond := &models.User{}
	findCond.ID = userId
	// 3. 执行更新
	tx := userRepo.db.WithContext(ctx).Model(&models.User{}).Where(findCond).Updates(updateModel)
	return tx.Error
}

/**
 * Update User Nickname
 */
func (userRepo *UserRepoImpl) UpdateNickname(
	ctx context.Context,
	userId int64,
	nickname string,
) error {
	// 1. 创建模型
	updateCond := &models.User{Nickname: nickname}
	findCond := &models.User{}
	findCond.ID = userId
	// 2. 执行更新
	tx := userRepo.db.WithContext(ctx).Model(&models.User{}).Where(findCond).Updates(updateCond)
	return tx.Error
}

/**
 * Update User Password
 */
func (userRepo *UserRepoImpl) UpdatePassword(
	ctx context.Context,
	userId int64,
	password string,
	newPassword string,
) error {
	// 1. 创建更新模型
	updateCond := &models.User{Password: newPassword}
	// 2. 创建查找模型
	findCond := &models.User{}
	findCond.ID = userId
	// 3. 查找用户记录
	user, err := userRepo.FindById(ctx, userId)
	if err != nil {
		return err
	}
	// 3. 对比旧密码
	isMatch := userRepo.PasswordCompare([]byte(password), []byte(user.Password))
	if !isMatch {
		return errors.New("旧密码错误")
	}
	// 4. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	updateCond.Password = string(hashedPassword)
	// 5. 执行更新
	tx := userRepo.db.WithContext(ctx).Model(&models.User{}).Where(findCond).Updates(updateCond)
	return tx.Error
}

/**
 * Compare Password
 */
func (userRepo *UserRepoImpl) PasswordCompare(
	password, encryptedPassword []byte,
) bool {
	return bcrypt.CompareHashAndPassword(encryptedPassword, password) == nil
}
