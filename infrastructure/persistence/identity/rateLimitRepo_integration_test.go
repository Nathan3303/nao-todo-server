//go:build integration

package identity

import (
	"context"
	"testing"
	"time"
)

// TestAllowEnforcesRedisQuota ① Redis 正常：按原配额生效（第 limit+1 次拒绝）
func TestAllowEnforcesRedisQuota(t *testing.T) {
	repo := NewRateLimitRepo(newTestRedisClient(t), NewLocalFallback())
	ctx := context.Background()
	key := "rate_limit:integration-quota:auth-signin"

	for i := int64(0); i < 3; i++ {
		ok, err := repo.Allow(ctx, key, 3)
		if err != nil {
			t.Fatalf("Allow 返回错误: %v", err)
		}
		if !ok {
			t.Fatalf("第 %d 次请求应放行（配额 3）", i+1)
		}
	}
	ok, err := repo.Allow(ctx, key, 3)
	if err != nil {
		t.Fatalf("Allow 返回错误: %v", err)
	}
	if ok {
		t.Fatal("第 4 次请求应被拒绝（配额 3，固定 60 秒窗口）")
	}
}

// TestAllowReturnsToRedisAfterRecovery ①/② Redis 恢复后自动回退到 Redis：
// 降级期间本地桶已耗尽，Redis 恢复后同一键重新按 Redis 配额放行。
func TestAllowReturnsToRedisAfterRecovery(t *testing.T) {
	healthy := newTestRedisClient(t)
	local := NewLocalFallback()
	repo := &RateLimitRepoImpl{rds: newDownRedisClient(t), local: local}
	ctx := context.Background()
	key := "rate_limit:integration-recovery:auth-signin"

	// 配额 8 ⇒ 降级本地桶 max(2,5)=5
	for i := 0; i < 5; i++ {
		if ok, _ := repo.Allow(ctx, key, 8); !ok {
			t.Fatalf("降级期第 %d 次应放行", i+1)
		}
	}
	if ok, _ := repo.Allow(ctx, key, 8); ok {
		t.Fatal("降级期第 6 次应被本地桶拒绝")
	}

	// Redis 恢复：下一次请求直接走 Redis，按原配额（8）放行
	repo.rds = healthy
	if ok, _ := repo.Allow(ctx, key, 8); !ok {
		t.Fatal("Redis 恢复后应自动回退到 Redis 并放行")
	}
	for i := 0; i < 7; i++ {
		if ok, _ := repo.Allow(ctx, key, 8); !ok {
			t.Fatalf("Redis 恢复后第 %d 次（总 8 次配额的范围内）应放行", i+2)
		}
	}
	if ok, _ := repo.Allow(ctx, key, 8); ok {
		t.Fatal("Redis 恢复后第 9 次应按 Redis 配额拒绝")
	}
	if got := local.Count(key); got != 6 {
		t.Fatalf("降级期本地计数 = %d, want 6（恢复后不再累加）", got)
	}
}

// TestAllowDegradedFallsBackToLocalBucketAndLogs ② Redis 故障：本地桶生效 + 每次降级一行日志。
func TestAllowDegradedFallsBackToLocalBucketAndLogs(t *testing.T) {
	logs := captureLogs(t)
	repo := NewRateLimitRepo(newDownRedisClient(t), NewLocalFallback())
	ctx := context.Background()
	key := "rate_limit:integration-degraded:auth-signin"

	allowed := 0
	for i := 0; i < 6; i++ {
		if ok, _ := repo.Allow(ctx, key, 8); ok {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("降级放行数 = %d, want 5", allowed)
	}
	entries := logEntries(t, logs)
	if findLogEntry(entries, "ratelimit.degraded") == nil {
		t.Fatal("降级必须记录 ratelimit.degraded")
	}
}

// TestAllowQuotaWindowExpires Redis 固定窗口到期后恢复放行
func TestAllowQuotaWindowExpires(t *testing.T) {
	rds := newTestRedisClient(t)
	repo := NewRateLimitRepo(rds, NewLocalFallback())
	ctx := context.Background()
	key := "rate_limit:integration-window:auth-signin"

	if ok, _ := repo.Allow(ctx, key, 1); !ok {
		t.Fatal("首次请求应放行")
	}
	if ok, _ := repo.Allow(ctx, key, 1); ok {
		t.Fatal("第 2 次应被拒绝")
	}
	if err := rds.Del(ctx, key).Err(); err != nil {
		t.Fatalf("清理键失败: %v", err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if ok, _ := repo.Allow(ctx, key, 1); ok {
			return
		}
	}
	t.Fatal("窗口重置后应恢复放行")
}
