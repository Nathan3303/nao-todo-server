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
	// FindByUserIdAndToken 按 (user_id, token) 查会话。
	// 「查不到」（无此行 / 软删）返回 (nil, nil)；查询失败（DB/连接类）返回 (nil, err)。
	// DEF-143：两者必须可区分，否则基础设施故障会被折叠成凭证失效（10041）。
	FindByUserIdAndToken(
		ctx context.Context,
		userId types.UserID,
		token string,
	) (*entities.UserSession, error)
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
