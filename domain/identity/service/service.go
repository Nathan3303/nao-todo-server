package service

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
)

// IdentityDomain 身份认证领域接口
type IdentityDomain interface {
	CreateSession(ctx context.Context, userId int64, token string) error
	FindSessionByUserIdAndToken(
		ctx context.Context,
		userId int64,
		token string,
	) (*entities.UserSession, error)
	DeleteSession(ctx context.Context, sessionEntity *entities.UserSession) error
	GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error)
	ParseJWT(ctx context.Context, token string) (int64, error)
	CheckRateLimit(ctx context.Context, key string, limit int8) error
}

type identityDomainImpl struct {
	jwtRepo         repositories.JWT
	userSessionRepo repositories.UserSession
	rateLimitRepo   repositories.RateLimit
}
