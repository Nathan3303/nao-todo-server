package repositories

import (
	"context"
	"time"

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
	// DeleteByUserIdExceptToken 删除除 keepToken 外的所有会话，返回实际撤销条数
	DeleteByUserIdExceptToken(
		ctx context.Context,
		userId types.UserID,
		keepToken string,
	) (int64, error)
	// DeleteExpiredBefore 硬删除过期时间早于 cutoff 且未登出的会话，最多删除 limit 行
	DeleteExpiredBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	// DeleteRevokedBefore 硬删除登出时间早于 cutoff 的会话，最多删除 limit 行
	DeleteRevokedBefore(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	IsSessionValid(ctx context.Context, userId types.UserID, token string) bool
	Ip2Region(ip string) (string, error)
}
