package identity

import (
	"context"
	"naotodoserver/domain/identity/repositories"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type RateLimitRepoImpl struct {
	rds *redis.Client
}

func NewRateLimitRepo(rds *redis.Client) repositories.RateLimit {
	return &RateLimitRepoImpl{
		rds: rds,
	}
}

func (rateLimitRepo *RateLimitRepoImpl) Get(ctx context.Context, key string) int8 {
	val, err := rateLimitRepo.rds.Get(ctx, key).Result()
	var count int8
	switch err {
	case redis.Nil:
		// key 不存在
		count = -1
	default:
		// key 存在时解析真实计数；其余错误（如 Redis 不可用）解析失败归 0，不误拦请求
		parsed, _ := strconv.ParseInt(val, 10, 8)
		count = int8(parsed)
	}
	return count
}

func (rateLimitRepo *RateLimitRepoImpl) Incr(ctx context.Context, key string) error {
	// 原子递增并设置过期时间（1 分钟）
	_, err := rateLimitRepo.rds.Incr(ctx, key).Result()
	if err != nil {
		// Redis 不可用时放行（fail-open），避免限流服务故障导致全站请求被拒
		return nil
	}
	// 设置 TTL（只在首次设置时生效，避免覆盖已有 TTL）
	rateLimitRepo.rds.ExpireNX(ctx, key, time.Minute)
	return nil
}
