package auth

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
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

/*
 * Create User
 */
func (ur *userRepoImpl) Create(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	// 1. 获取客户端信息
	clientInfo := iCtx.GetClientInfo(ctx)
	// 2. 实体转换
	userModel := UserEntity2Model(userEntity)
	userModel.CreatedFrom = clientInfo.IPRegion + " " + clientInfo.DeviceType
	// 3. 创建用户
	tx := ur.db.WithContext(ctx).Model(&models.User{}).Create(userModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 模型转换并返回
	return UserModel2Entity(userModel), nil
}

/*
 * Find User By Email
 */
func (ur *userRepoImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	// 1. 创建结果模型
	user := &models.User{}
	// 2. 创建查找模型
	findCond := &models.User{Email: email}
	// 3. 执行查找
	ur.db.WithContext(ctx).Model(&models.User{}).Where(findCond).First(&user)
	// tx := ur.db.WithContext(ctx).Model(&models.User{}).Where(findCond).First(&user)
	// if tx.Error != nil {
	// 	return nil, tx.Error
	// }
	// 4. 模型转换并返回
	return UserModel2Entity(user), nil
}

/*
 * Find User By Id
 */
func (ur *userRepoImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	// 1. 创建结果模型
	user := &models.User{}
	// 2. 创建查找模型
	findCond := &models.User{}
	findCond.ID = id
	// 3. 执行查找
	ur.db.WithContext(ctx).Model(&models.User{}).Where(findCond).First(&user)
	// if tx.Error != nil {
	// 	return nil, tx.Error
	// }
	// 4. 模型转换并返回
	return UserModel2Entity(user), nil
}

/*
 * Compare Password
 */
func (u *userRepoImpl) PasswordCompare(password []byte, encryptedPassword []byte) bool {
	return bcrypt.CompareHashAndPassword(encryptedPassword, password) == nil
}

/*
 * Get Rate Limit
 */
func (u *userRepoImpl) GetRateLimit(ctx context.Context, key string) (int64, error) {
	return u.rds.Get(ctx, key).Int64()
}

/*
 * Increment Rate Limit
 */
func (u *userRepoImpl) IncrementRateLimit(ctx context.Context, key string) error {
	return u.rds.Incr(ctx, key).Err()
}
