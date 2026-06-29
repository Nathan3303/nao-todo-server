package repositories

import (
	"context"
	"naotodoserver/domain/identity/entities"
)

// UserSession 用户会话仓库接口
type UserSession interface {
	Create(ctx context.Context, sessionEntity *entities.UserSession) error
	FindByUserIdAndToken(ctx context.Context, userId int64, token string) *entities.UserSession
	UpdateToken(ctx context.Context, sessionEntity *entities.UserSession) error
	Delete(ctx context.Context, userId int64, token string) error
	IsSessionValid(ctx context.Context, userId int64, token string) bool
	Ip2Region(ip string) (string, error)
}
