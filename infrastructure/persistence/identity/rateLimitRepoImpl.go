package identity

import (
	"context"

	"naotodoserver/domain/identity/repositories"
	"naotodoserver/infrastructure/logging"

	"github.com/go-redis/redis/v8"
)

// localFallbackRatio 降级时本地桶容量为 Redis 配额的 1/4
const localFallbackRatio = 4

// localFallbackMinLimit 降级时本地桶容量下限（每分钟）
const localFallbackMinLimit = 5

type RateLimitRepoImpl struct {
	rds   *redis.Client
	local *LocalFallback
}

// NewRateLimitRepo 创建限流仓储
// @param rds Redis 客户端
// @param local 进程内兜底计数器（Redis 不可用时使用）
func NewRateLimitRepo(rds *redis.Client, local *LocalFallback) repositories.RateLimit {
	return &RateLimitRepoImpl{
		rds:   rds,
		local: local,
	}
}

// allowScript 原子执行"检查 + 计数"：
// - 窗口内计数达到 limit 时返回 -1（拒绝），超限请求不计数
// - 否则计数 +1 并返回当前计数（放行）
// - 首次计数时设置窗口过期时间（60 秒）
const allowScript = `
local c = redis.call('GET', KEYS[1])
if c and tonumber(c) >= tonumber(ARGV[1]) then
  return -1
end
local n = redis.call('INCR', KEYS[1])
if n == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[2])
end
return n
`

// Allow 原子地检查并计数（避免 Get + Incr 两步操作的并发竞态）
// Redis 正常时按其固定窗口裁决；Redis 不可用时降级到进程内本地桶
// （fail-open 的业务可用性 + 有限的节流能力，SEC-10），每次降级记一行结构化日志。
// @param ctx 上下文
// @param key 计数键（客户端 IP 或用户维度）
// @param limit 每分钟配额（Redis 侧）
// @return 是否放行
func (rateLimitRepo *RateLimitRepoImpl) Allow(
	ctx context.Context,
	key string,
	limit int64,
) (bool, error) {
	val, err := rateLimitRepo.rds.Eval(
		ctx,
		allowScript,
		[]string{key},
		limit,
		localWindowSeconds,
	).Result()
	if err != nil {
		fallbackLimit := localFallbackLimit(limit)
		logRateLimitDegraded(key, err, limit, fallbackLimit)
		return rateLimitRepo.local.Incr(key, localWindowDuration) <= fallbackLimit, nil
	}
	count, ok := val.(int64)
	if !ok {
		// 返回类型异常（非预期）：保持既有 fail-open 语义
		logging.Warnf("rate limit unexpected reply type: key=%s", keyTag(key))
		return true, nil
	}
	return count >= 0, nil
}

// localFallbackLimit 计算降级时的本地桶配额：max(配额/4, 5)
// @param limit Redis 侧配额
// @return 本地桶配额
func localFallbackLimit(limit int64) int64 {
	quota := limit / localFallbackRatio
	if quota < localFallbackMinLimit {
		return localFallbackMinLimit
	}
	return quota
}
