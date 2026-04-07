package service

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/domain/auth/valueobjects"
)

// AuthDomain 认证领域接口
type AuthDomain interface {
	// CreateUser 创建用户
	CreateUser(
		ctx context.Context,
		createUserValueObject *valueobjects.CreateUser,
	) (*entities.User, error)

	// CreateSession 创建会话
	CreateSession(ctx context.Context, userId int64, token string) error

	// FindUserByEmail 根据邮箱查找用户
	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindUserById 根据用户ID查找用户
	FindUserById(ctx context.Context, id int64) (*entities.User, error)

	// FindSessionByUserIdAndToken 根据用户ID和会话令牌查找会话
	FindSessionByUserIdAndToken(
		ctx context.Context,
		userId int64,
		token string,
	) (*entities.Session, error)

	// UpdateSessionToken 更新会话令牌
	UpdateSessionToken(ctx context.Context, sessionEntity *entities.Session) error

	// DeleteSession 删除会话
	DeleteSession(ctx context.Context, sessionEntity *entities.Session) error

	// GenerateJWT 生成JWT
	GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error)

	// ParseJWT 解析JWT
	ParseJWT(ctx context.Context, token string) (int64, error)

	// PasswordCompare 密码比较
	PasswordCompare(ctx context.Context, password, encryptedPassword []byte) bool

	// CheckRateLimit 检查速率限制
	CheckRateLimit(ctx context.Context, key string, limit int8) error
}

// authDomainImpl 认证领域实现
type authDomainImpl struct {
	jwtRepo       repositories.JWT
	userRepo      repositories.User
	sessionRepo   repositories.Session
	rateLimitRepo repositories.RateLimit
}
