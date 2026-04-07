package repositories

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/valueobjects"
)

// User 用户仓库接口
type User interface {
	// Create 创建用户
	Create(
		ctx context.Context,
		createUserValueObject *valueobjects.CreateUser,
	) (*entities.User, error)

	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindById 根据用户ID查找用户
	FindById(ctx context.Context, id int64) (*entities.User, error)

	// PasswordCompare 密码比较
	PasswordCompare(password, encryptedPassword []byte) bool

	// GetRateLimit 获取速率限制
	GetRateLimit(ctx context.Context, key string) (int64, error)

	// IncrementRateLimit 增加速率限制
	IncrementRateLimit(ctx context.Context, key string) error
}
