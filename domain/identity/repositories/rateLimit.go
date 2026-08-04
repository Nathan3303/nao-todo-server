package repositories

import "context"

type RateLimit interface {
	// Allow 原子地检查并计数（单位窗口），返回是否放行；Redis 不可用时由实现决定是否放行
	Allow(ctx context.Context, key string, limit int64) (bool, error)
}
