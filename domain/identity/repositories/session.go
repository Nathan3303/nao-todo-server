package repositories

import (
	"context"
	"naotodoserver/domain/identity/entities"
)

// Session 会话仓库接口
type Session interface {
	Create(ctx context.Context, sessionEntity *entities.Session) error
	FindByUserIdAndToken(ctx context.Context, userId int64, token string) *entities.Session
	UpdateToken(ctx context.Context, sessionEntity *entities.Session) error
	Delete(ctx context.Context, userId int64, token string) error
	IsSessionValid(ctx context.Context, userId int64, token string) bool
	Ip2Region(ip string) (string, error)
}
