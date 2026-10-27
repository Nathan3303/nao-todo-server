//go:build integration

package identity

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v8"
)

// integrationRedisDB 集成测试专用逻辑库，避免污染开发库键空间
const integrationRedisDB = 15

// newTestRedisClient 连接真实 Redis：
// NAO_TEST_REDIS_ADDR / NAO_TEST_REDIS_PASSWORD / NAO_TEST_REDIS_DB（默认 127.0.0.1:6379 / 空 / 15）
func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	db := integrationRedisDB
	if raw := os.Getenv("NAO_TEST_REDIS_DB"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			t.Fatalf("NAO_TEST_REDIS_DB 非法: %v", err)
		}
		db = parsed
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("NAO_TEST_REDIS_PASSWORD"),
		DB:       db,
	})
	t.Cleanup(func() {
		_ = client.FlushDB(context.Background()).Err()
		_ = client.Close()
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("连接测试 Redis 失败（需先启动测试 Redis）: %v", err)
	}
	// 专用逻辑库：每个用例前清空，避免相互干扰
	if err := client.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("清空测试 Redis 库失败: %v", err)
	}
	return client
}
