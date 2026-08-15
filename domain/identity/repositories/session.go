package repositories

import (
	"context"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/types"
)

// UserSession 用户会话仓库接口
type UserSession interface {
	Create(ctx context.Context, sessionEntity *entities.UserSession) error
	FindByUserIdAndToken(
		ctx context.Context,
		userId types.UserID,
		token string,
	) *entities.UserSession
	FindByUserId(ctx context.Context, userId types.UserID) ([]*entities.UserSession, error)
	UpdateToken(ctx context.Context, sessionEntity *entities.UserSession, oldToken string) error
	Delete(ctx context.Context, userId types.UserID, token string) error
	DeleteById(ctx context.Context, userId types.UserID, sessionId int64) error
	DeleteByUserId(ctx context.Context, userId types.UserID) error
	DeleteByUserIdExceptToken(ctx context.Context, userId types.UserID, keepToken string) error
	IsSessionValid(ctx context.Context, userId types.UserID, token string) bool
	Ip2Region(ip string) (string, error)
}
