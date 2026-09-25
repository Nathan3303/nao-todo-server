package identity

import (
	"context"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/infrastructure/logging"

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
// Redis 不可用时放行（fail-open），避免限流服务故障导致全站请求被拒
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
		60,
	).Result()
	if err != nil {
		// Redis 不可用时放行（fail-open），避免限流服务故障导致全站请求被拒；记录告警以便发现限流失效
		logging.Warnf("rate limit skipped (redis unavailable): key=%s err=%v", key, err)
		return true, nil
	}
	count, ok := val.(int64)
	if !ok {
		return true, nil
	}
	return count >= 0, nil
}
