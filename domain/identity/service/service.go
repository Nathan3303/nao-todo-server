package service

import (
	"context"

	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/types"
)

// IdentityDomain 身份认证领域接口
type IdentityDomain interface {
	CreateSession(ctx context.Context, userId types.UserID, token string) error
	FindSessionByUserIdAndToken(
		ctx context.Context,
		userId types.UserID,
		token string,
	) (*entities.UserSession, error)
	DeleteSession(ctx context.Context, sessionEntity *entities.UserSession) error
	ListSessions(ctx context.Context, userId types.UserID) ([]*entities.UserSession, error)
	DeleteSessionById(ctx context.Context, userId types.UserID, sessionId int64) error
	DeleteOtherSessions(ctx context.Context, userId types.UserID, keepToken string) error
	GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error)
	ParseJWT(ctx context.Context, token string) (types.UserID, error)
	IsJWTExpired(ctx context.Context, token string) bool
	CheckRateLimit(ctx context.Context, key string, limit int64) error
	// IsLoginBlocked 账号是否因连续登录失败被暂停（SEC-04，轻量查询，不查库）
	IsLoginBlocked(ctx context.Context, email string) bool
	// RecordLoginFailure 记录一次登录失败；达到阈值时暂停该账号登录
	RecordLoginFailure(ctx context.Context, email string)
	// ClearLoginFailures 登录成功后清除账号的失败计数与暂停状态
	ClearLoginFailures(ctx context.Context, email string)
}

type identityDomainImpl struct {
	jwtRepo         repositories.JWT
	userSessionRepo repositories.UserSession
	rateLimitRepo   repositories.RateLimit
	loginThrottle   repositories.LoginThrottle
}
