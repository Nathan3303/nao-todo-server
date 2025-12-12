package dbs

import (
	"context"
	"fmt"
	"log"
	"naotodoserver/conf"

	"github.com/go-redis/redis/v8"
)

var RdsCli *redis.Client

// InitRedis 初始化 Redis 客户端
func InitRedis() {
	// 1. 创建 Redis 客户端
	ctx := context.Background()
	redisConfig := conf.Conf.Redis
	client := redis.NewClient(&redis.Options{
		Addr: redisConfig.Host + ":" + redisConfig.Port,
		// Password: redisConfig.Password,
		Password: "",
		DB:       redisConfig.DB,
	})

	// 2. 测试连接
	err := client.Ping(ctx).Err()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("✅ Connected to Redis!")

	// 3. 赋值客户端
	RdsCli = client
}
