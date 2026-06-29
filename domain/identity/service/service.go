package service

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
)

// IdentityDomain 身份认证与用户管理领域接口
type IdentityDomain interface {
	// --- 认证相关 ---
	CreateUser(ctx context.Context, vo *valueobjects.CreateUser) (*entities.User, error)
	CreateSession(ctx context.Context, userId int64, token string) error
	FindSessionByUserIdAndToken(
		ctx context.Context,
		userId int64,
		token string,
	) (*entities.UserSession, error)
	UpdateSessionToken(ctx context.Context, sessionEntity *entities.UserSession) error
	DeleteSession(ctx context.Context, sessionEntity *entities.UserSession) error
	GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error)
	ParseJWT(ctx context.Context, token string) (int64, error)
	CheckRateLimit(ctx context.Context, key string, limit int8) error

	// --- 用户管理 ---
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindById(ctx context.Context, id int64) (*entities.User, error)
	UpdateNickname(ctx context.Context, userId int64, nickname string) error
	UpdateAvatar(ctx context.Context, userId int64, avatar string) error
	UpdatePassword(ctx context.Context, userId int64, password, newPassword string) error
	Deactive(ctx context.Context, userId int64) error
	Active(ctx context.Context, userId int64) error
	PasswordCompare(password, encryptedPassword []byte) bool
	GetConfig(ctx context.Context, userId int64) (*entities.UserConfig, error)
	UpdateConfig(ctx context.Context, userId int64, appearance string) error
	DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) (int64, error)
}

type identityDomainImpl struct {
	jwtRepo         repositories.JWT
	userRepo        repositories.User
	userSessionRepo repositories.UserSession
	rateLimitRepo   repositories.RateLimit
}
