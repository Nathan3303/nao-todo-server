package repositories

import (
	"context"
	"naotodoserver/domain/auth/entities"
)

// Session 会话仓库接口
type Session interface {
	// Create 创建会话
	Create(ctx context.Context, sessionEntity *entities.Session) error

	// FindByUserIdAndToken 根据用户ID和会话令牌查找会话
	FindByUserIdAndToken(ctx context.Context, userId int64, token string) *entities.Session

	// UpdateToken 更新会话令牌
	UpdateToken(ctx context.Context, sessionEntity *entities.Session) error

	// Delete 删除会话
	Delete(ctx context.Context, userId int64, token string) error

	// Ip2Region IP转区域
	// @param ip IP地址
	// @return 区域名称
	// @return error 错误
	Ip2Region(ip string) (string, error)
}
