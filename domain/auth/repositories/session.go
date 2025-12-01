package repositories

import (
	"context"
	"naotodoserver/domain/auth/entities"
)

type Session interface {
	Create(ctx context.Context, sessionEntity *entities.Session) error
	FindByUserIdAndToken(ctx context.Context, userId int64, token string) *entities.Session
	UpdateToken(ctx context.Context, sessionEntity *entities.Session) error
	Delete(ctx context.Context, userId int64, token string) error
}
