package repositories

import (
	"context"
	"naotodoserver/domain/auth/entities"
)

type User interface {
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindById(ctx context.Context, id int64) (*entities.User, error)
	PasswordCompare(password, encryptedPassword []byte) bool
	GetRateLimit(ctx context.Context, key string) (int64, error)
	IncrementRateLimit(ctx context.Context, key string) error
}
