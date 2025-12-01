package service

import (
	"context"
	"naotodoserver/domain/auth/entities"
)

type authDomain interface {
	CreateUser(ctx context.Context, userEntity *entities.User) (*entities.User, error)
	CreateSession(ctx context.Context, userId int64, token string) error

	FindUserByEmail(ctx context.Context, email string) (*entities.User, error)
	FindSessionByUserIdAndToken(
		ctx context.Context, userId int64, token string,
	) (*entities.Session, error)

	UpdateSessionToken(ctx context.Context, sessionEntity *entities.Session) error
	DeleteSession(ctx context.Context, sessionEntity *entities.Session) error

	GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error)
	ValidateJWT(ctx context.Context, token string) bool

	PasswordCompare(ctx context.Context, password, encryptedPassword []byte) bool
}
