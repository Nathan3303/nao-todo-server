package service

import (
	"context"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/infrastructure/auth"
)

type AuthDomain struct {
	jwt         *auth.JWTServiceImpl
	userRepo    *repositories.User
	sessionRepo *repositories.Session
}

func GetAuthDomainImpl(
	jwt *auth.JWTServiceImpl,
	userRepo *repositories.User,
	sessionRepo *repositories.Session,
) authDomain {
	return &AuthDomain{
		jwt:         jwt,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// CreateSession implements authDomain.
func (a *AuthDomain) CreateSession(ctx context.Context, userId int64, token string) error {
	panic("unimplemented")
}

// CreateUser implements authDomain.
func (a *AuthDomain) CreateUser(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	panic("unimplemented")
}

// DeleteSession implements authDomain.
func (a *AuthDomain) DeleteSession(ctx context.Context, sessionEntity *entities.Session) error {
	panic("unimplemented")
}

// FindSessionByUserIdAndToken implements authDomain.
func (a *AuthDomain) FindSessionByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) (*entities.Session, error) {
	panic("unimplemented")
}

// FindUserByEmail implements authDomain.
func (a *AuthDomain) FindUserByEmail(
	ctx context.Context,
	email string,
) (*entities.User, error) {
	panic("unimplemented")
}

// GenerateJWT implements authDomain.
func (a *AuthDomain) GenerateJWT(
	ctx context.Context,
	userEntity *entities.User,
) (string, error) {
	panic("unimplemented")
}

// PasswordCompare implements authDomain.
func (a *AuthDomain) PasswordCompare(
	ctx context.Context,
	password []byte,
	encryptedPassword []byte,
) bool {
	panic("unimplemented")
}

// UpdateSessionToken implements authDomain.
func (a *AuthDomain) UpdateSessionToken(
	ctx context.Context,
	sessionEntity *entities.Session,
) error {
	panic("unimplemented")
}

// ValidateJWT implements authDomain.
func (a *AuthDomain) ValidateJWT(ctx context.Context, token string) bool {
	panic("unimplemented")
}
