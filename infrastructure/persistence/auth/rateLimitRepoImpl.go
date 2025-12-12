package auth

import (
	"context"
	"errors"
	"naotodoserver/domain/auth/repositories"
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
	var count int64
	if err == redis.Nil {
		count = -1
	} else if err != nil {
		count = 0
	} else {
		count, _ = strconv.ParseInt(val, 10, 64)
	}
	return int8(count)
}

func (rateLimitRepo *RateLimitRepoImpl) Incr(ctx context.Context, key string) error {
	// 原子递增并设置过期时间（1 小时）
	_, err := rateLimitRepo.rds.Incr(ctx, key).Result()
	if err != nil {
		return errors.New("限流阈值设置失败" + err.Error())
	}
	// 设置 TTL（只在首次设置时生效，避免覆盖已有 TTL）
	rateLimitRepo.rds.ExpireNX(ctx, key, time.Minute)
	return nil
}
