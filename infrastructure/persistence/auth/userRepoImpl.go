package auth

import (
	"context"
	"errors"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/domain/auth/valueobjects"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/persistence/models"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepoImpl struct {
	db  *gorm.DB
	rds *redis.Client
}

func NewUserRepo(db *gorm.DB, rds *redis.Client) repositories.User {
	return &userRepoImpl{
		db:  db,
		rds: rds,
	}
}

// Create 创建用户
// @param ctx 上下文
// @param userEntity 用户实体
// @return 用户实体
// @return error 错误
func (ur *userRepoImpl) Create(
	ctx context.Context,
	createUserValueObject *valueobjects.CreateUser,
) (*entities.User, error) {
	// 1. 获取客户端信息
	clientInfo := iCtx.GetClientInfo(ctx)
	// 2. 实体转换
	userModel := CreateUserValueObjectToModel(createUserValueObject)
	userModel.CreatedFrom = clientInfo.IPRegion + " " + clientInfo.DeviceType
	// 3. 创建用户
	tx := ur.db.WithContext(ctx).Model(&models.User{}).Create(userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 创建用户默认配置
	userConfig := &models.UserConfig{UserId: userModel.ID, Appearance: "auto"}
	ur.db.WithContext(ctx).Create(userConfig)
	// 5. 模型转换并返回
	return UserModel2Entity(userModel), nil
}

// FindByEmail 根据邮箱查找用户
// @param ctx 上下文
// @param email 邮箱
// @return 用户实体
// @return error 错误
func (ur *userRepoImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	// 1. 创建结果模型
	user := &models.User{}
	// 2. 创建查找模型
	findCond := &models.User{Email: email}
	// 3. 执行查找
	ur.db.WithContext(ctx).Model(&models.User{}).Where(findCond).First(&user)
	if user.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	// 4. 模型转换并返回
	return UserModel2Entity(user), nil
}

// FindById 根据ID查找用户
// @param ctx 上下文
// @param id 用户ID
// @return 用户实体
// @return error 错误
func (ur *userRepoImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	// 1. 创建结果模型
	user := &models.User{}
	// 2. 创建查找模型
	findCond := &models.User{}
	findCond.ID = id
	// 3. 执行查找
	ur.db.WithContext(ctx).Model(&models.User{}).Where(findCond).First(&user)
	if user.ID == 0 {
		return nil, errors.New("用户不存在")
	}
	// 4. 模型转换并返回
	return UserModel2Entity(user), nil
}

// PasswordCompare 比对密码
// @param password 明文密码
// @param encryptedPassword 加密后的密码
// @return 是否匹配
func (u *userRepoImpl) PasswordCompare(password []byte, encryptedPassword []byte) bool {
	return bcrypt.CompareHashAndPassword(encryptedPassword, password) == nil
}

// GetRateLimit 获取速率限制
// @param ctx 上下文
// @param key 速率限制键
// @return 速率限制值
// @return error 错误
func (u *userRepoImpl) GetRateLimit(ctx context.Context, key string) (int64, error) {
	return u.rds.Get(ctx, key).Int64()
}

// IncrementRateLimit 增加速率限制
// @param ctx 上下文
// @param key 速率限制键
// @return error 错误
func (u *userRepoImpl) IncrementRateLimit(ctx context.Context, key string) error {
	return u.rds.Incr(ctx, key).Err()
}
