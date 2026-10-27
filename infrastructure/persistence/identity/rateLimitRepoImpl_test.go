package identity

import (
	"context"
	"strings"
	"testing"
)

// TestAllowFallsBackToLocalBucketWhenRedisDown Redis 故障 ⇒ 本地桶按 max(配额/4, 5) 生效，
// 超限仍拒绝（不改 10051 语义）。
func TestAllowFallsBackToLocalBucketWhenRedisDown(t *testing.T) {
	logs := captureLogs(t)
	local := NewLocalFallback()
	repo := &RateLimitRepoImpl{rds: newDownRedisClient(t), local: local}
	ctx := context.Background()
	key := "203.0.113.7:auth-signin"

	// 配额 40 ⇒ 本地桶 40/4 = 10
	allowed := 0
	for i := 0; i < 12; i++ {
		ok, err := repo.Allow(ctx, key, 40)
		if err != nil {
			t.Fatalf("降级路径不应返回错误: %v", err)
		}
		if ok {
			allowed++
			continue
		}
		if allowed != 10 {
			t.Fatalf("第 %d 次请求被拒时放行数 = %d, want 10", i+1, allowed)
		}
	}
	if allowed != 10 {
		t.Fatalf("本地桶放行数 = %d, want 10", allowed)
	}

	// 每次降级一行结构化日志，且不含 IP 全量
	entries := logEntries(t, logs)
	degraded := 0
	for _, entry := range entries {
		if entry["event"] != "ratelimit.degraded" {
			continue
		}
		degraded++
		if entry["reason"] != "redis_unavailable" {
			t.Fatalf("reason = %v, want redis_unavailable", entry["reason"])
		}
		if _, ok := entry["key_tag"].(string); !ok {
			t.Fatalf("key_tag 缺失: %v", entry)
		}
	}
	if degraded != 12 {
		t.Fatalf("降级日志行数 = %d, want 12（每次降级一行）", degraded)
	}
	if strings.Contains(logs.String(), "203.0.113.7") {
		t.Fatal("降级日志不得包含客户端 IP 全量")
	}
}

// TestAllowLocalFallbackFloor Redis 故障时的本地桶容量下限为 5/分钟
// （signin 配额 8 ⇒ max(2, 5) = 5，保证兜底仍能拦）。
func TestAllowLocalFallbackFloor(t *testing.T) {
	captureLogs(t)
	repo := &RateLimitRepoImpl{rds: newDownRedisClient(t), local: NewLocalFallback()}
	ctx := context.Background()

	allowed := 0
	for i := 0; i < 8; i++ {
		ok, err := repo.Allow(ctx, "198.51.100.9:auth-signin", 8)
		if err != nil {
			t.Fatalf("降级路径不应返回错误: %v", err)
		}
		if ok {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("signin 降级放行数 = %d, want 5（下限生效）", allowed)
	}
	if localFallbackLimit(1) != 5 || localFallbackLimit(8) != 5 || localFallbackLimit(120) != 30 {
		t.Fatalf("localFallbackLimit 口径错误: %d/%d/%d",
			localFallbackLimit(1), localFallbackLimit(8), localFallbackLimit(120))
	}
}

// TestAllowLocalBucketIsPerKey 本地桶按 key 维度隔离（不同 IP / 不同桶互不拖累）
func TestAllowLocalBucketIsPerKey(t *testing.T) {
	captureLogs(t)
	repo := &RateLimitRepoImpl{rds: newDownRedisClient(t), local: NewLocalFallback()}
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if ok, _ := repo.Allow(ctx, "ip-a:auth-signin", 8); !ok {
			t.Fatalf("ip-a 第 %d 次应放行", i+1)
		}
	}
	if ok, _ := repo.Allow(ctx, "ip-a:auth-signin", 8); ok {
		t.Fatal("ip-a 超限后应拒绝")
	}
	if ok, _ := repo.Allow(ctx, "ip-b:auth-signin", 8); !ok {
		t.Fatal("ip-b 不应受 ip-a 影响")
	}
}
