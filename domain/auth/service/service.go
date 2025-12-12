package service

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
)

type AuthDomain interface {
	CreateUser(ctx context.Context, userEntity *entities.User) (*entities.User, error)
	CreateSession(ctx context.Context, userId int64, token string) error
	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)
	FindUserById(ctx context.Context, id int64) (*entities.User, error)
	FindSessionByUserIdAndToken(
		ctx context.Context,
		userId int64,
		token string,
	) (*entities.Session, error)
	UpdateSessionToken(ctx context.Context, sessionEntity *entities.Session) error
	DeleteSession(ctx context.Context, sessionEntity *entities.Session) error
	GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error)
	ParseJWT(ctx context.Context, token string) (int64, error)
	PasswordCompare(ctx context.Context, password, encryptedPassword []byte) bool
	CheckRateLimit(ctx context.Context, key string, limit int8) error
}

type authDomainImpl struct {
	jwtRepo       repositories.JWT
	userRepo      repositories.User
	sessionRepo   repositories.Session
	rateLimitRepo repositories.RateLimit
}
