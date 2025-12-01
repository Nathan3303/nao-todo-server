package repositories

import (
	"context"
	"naotodoserver/domain/auth/entities"
)

type User interface {
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	PasswordCompare(password, encryptedPassword []byte) bool
}
