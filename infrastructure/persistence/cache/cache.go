package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache 基于 Redis 的缓存辅助组件，提供 cache-aside 所需的 JSON 序列化能力
type Cache struct {
	rds *redis.Client
}

// NewCache 创建缓存辅助组件
func NewCache(rds *redis.Client) *Cache {
	return &Cache{
		rds: rds,
	}
}

// Get 从缓存读取并反序列化到 dest；命中返回 true，未命中或任何错误返回 false（视为未命中降级）
func (c *Cache) Get(ctx context.Context, key string, dest any) bool {
	val, err := c.rds.Get(ctx, key).Bytes()
	if err != nil {
		// redis.Nil 表示未命中，其余错误（含 Redis 不可用）统一降级为未命中
		return false
	}
	if err := json.Unmarshal(val, dest); err != nil {
		return false
	}
	return true
}

// Set 将 value JSON 序列化后写入缓存并设置 TTL；出错时静默忽略
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	c.rds.Set(ctx, key, data, ttl)
}

// Del 删除一个或多个 key；出错时静默忽略
func (c *Cache) Del(ctx context.Context, keys ...string) {
	if len(keys) == 0 {
		return
	}
	c.rds.Del(ctx, keys...)
}

// SessionKey 构造会话缓存 key
func SessionKey(userId int64, token string) string {
	return "cache:session:" + strconv.FormatInt(userId, 10) + ":" + token
}

// UserProfileKey 构造用户资料缓存 key
func UserProfileKey(userId int64) string {
	return "cache:user:profile:" + strconv.FormatInt(userId, 10)
}

// UserConfigKey 构造用户配置缓存 key
func UserConfigKey(userId int64) string {
	return "cache:user:config:" + strconv.FormatInt(userId, 10)
}

// ProjectListKey 构造项目列表缓存 key
func ProjectListKey(userId int64) string {
	return "cache:project:list:" + strconv.FormatInt(userId, 10)
}

// TagListKey 构造标签列表缓存 key
func TagListKey(userId int64) string {
	return "cache:tag:list:" + strconv.FormatInt(userId, 10)
}
