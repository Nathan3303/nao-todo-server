package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"naotodoserver/infrastructure/logging"
)

// keyTagLength 日志键标识长度（SHA-256 十六进制前 8 位）
const keyTagLength = 8

// keyTag 生成日志可用的键标识：SHA-256 前 8 位十六进制。
// 限流键含客户端 IP、账号键含邮箱哈希，日志中只保留不可逆摘要，
// 既不落 IP 全量也不落邮箱（SEC-10）。
// @param key 原始计数键
// @return 8 位十六进制摘要
func keyTag(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])[:keyTagLength]
}

// redisFailureReason 归一化 Redis 故障原因，避免把连接地址等细节打进结构化日志
// @param err Redis 调用错误
// @return 稳定原因码
func redisFailureReason(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, os.ErrDeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	default:
		return "redis_unavailable"
	}
}

// logRateLimitDegraded 记录一次限流降级（每降级一次一行结构化日志）
// @param key 计数键（仅以摘要形式落日志）
// @param err Redis 故障
// @param limit Redis 侧配额
// @param fallbackLimit 本地桶配额
func logRateLimitDegraded(key string, err error, limit int64, fallbackLimit int64) {
	logging.WarnWithFields(map[string]any{
		"event":          "ratelimit.degraded",
		"reason":         redisFailureReason(err),
		"key_tag":        keyTag(key),
		"limit":          limit,
		"fallback_limit": fallbackLimit,
	}, "rate limit degraded to local fallback")
}

// logLoginThrottleDegraded 记录一次账号节流降级（与 SEC-10 共用本地兜底）
// @param key 计数键（仅以摘要形式落日志）
// @param err Redis 故障
func logLoginThrottleDegraded(key string, err error) {
	logging.WarnWithFields(map[string]any{
		"event":   "login.throttle.degraded",
		"reason":  redisFailureReason(err),
		"key_tag": keyTag(key),
	}, "login throttle degraded to local fallback")
}

// localWindowSeconds 限流兜底窗口与 Redis 固定窗口保持一致（60 秒）
const localWindowSeconds = 60

// localWindowDuration 限流兜底窗口长度
const localWindowDuration = localWindowSeconds * time.Second
