package identity

import (
	"context"
	"time"

	"naotodoserver/domain/identity/repositories"

	"github.com/go-redis/redis/v8"
)

// 账号节流键前缀（Redis 键与本地兜底键共用；键体为邮箱哈希，不含明文）
const (
	loginFailKeyPrefix  = "login_fail:"
	loginBlockKeyPrefix = "login_block:"
)

// recordFailureScript 原子执行"计数 + 首次设置窗口过期"（毫秒精度 TTL）
const recordFailureScript = `
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return n
`

// LoginThrottleRepoImpl 账号维度登录失败计数与暂停（SEC-04）。
// Redis 不可用时退化为进程内计数（与 SEC-10 共用 LocalFallback），不阻断登录。
type LoginThrottleRepoImpl struct {
	rds   *redis.Client
	local *LocalFallback
}

// NewLoginThrottleRepo 创建账号节流仓储
// @param rds Redis 客户端
// @param local 进程内兜底计数器
func NewLoginThrottleRepo(rds *redis.Client, local *LocalFallback) repositories.LoginThrottle {
	return &LoginThrottleRepoImpl{
		rds:   rds,
		local: local,
	}
}

// RecordFailure 记录一次登录失败并返回窗口内累计失败次数
// @param accountKey 账号标识（哈希，不含明文）
// @param window 失败计数窗口
// @return 窗口内累计失败次数
func (repo *LoginThrottleRepoImpl) RecordFailure(
	ctx context.Context,
	accountKey string,
	window time.Duration,
) int64 {
	key := loginFailKeyPrefix + accountKey
	val, err := repo.rds.Eval(
		ctx,
		recordFailureScript,
		[]string{key},
		window.Milliseconds(),
	).Result()
	if err != nil {
		logLoginThrottleDegraded(key, err)
		return repo.local.Incr(key, window)
	}
	count, ok := val.(int64)
	if !ok {
		return repo.local.Incr(key, window)
	}
	return count
}

// Block 将账号置为暂停登录状态（duration 到期自动解除）
// @param accountKey 账号标识（哈希，不含明文）
// @param duration 暂停时长
func (repo *LoginThrottleRepoImpl) Block(
	ctx context.Context,
	accountKey string,
	duration time.Duration,
) {
	key := loginBlockKeyPrefix + accountKey
	if err := repo.rds.Set(ctx, key, 1, duration).Err(); err != nil {
		logLoginThrottleDegraded(key, err)
		repo.local.Incr(key, duration)
	}
}

// IsBlocked 轻量查询账号是否处于暂停登录状态（单次查询，不查库）
// @param accountKey 账号标识（哈希，不含明文）
// @return 是否暂停登录
func (repo *LoginThrottleRepoImpl) IsBlocked(
	ctx context.Context,
	accountKey string,
) bool {
	key := loginBlockKeyPrefix + accountKey
	exists, err := repo.rds.Exists(ctx, key).Result()
	if err != nil {
		logLoginThrottleDegraded(key, err)
		return repo.local.Count(key) > 0
	}
	return exists > 0
}

// Reset 清除账号的失败计数与暂停状态（登录成功后调用）
// @param accountKey 账号标识（哈希，不含明文）
func (repo *LoginThrottleRepoImpl) Reset(ctx context.Context, accountKey string) {
	failKey := loginFailKeyPrefix + accountKey
	blockKey := loginBlockKeyPrefix + accountKey
	if err := repo.rds.Del(ctx, failKey, blockKey).Err(); err != nil {
		logLoginThrottleDegraded(failKey, err)
	}
	repo.local.Reset(failKey)
	repo.local.Reset(blockKey)
}
