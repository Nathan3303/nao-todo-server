//go:build integration

package identity

import (
	"context"
	"testing"
	"time"
)

// TestLoginThrottleRedisBehaviour ③/④/⑤ Redis 正常：失败计数 → 暂停 → 成功清零 → 到期解除
func TestLoginThrottleRedisBehaviour(t *testing.T) {
	rds := newTestRedisClient(t)
	repo := NewLoginThrottleRepo(rds, NewLocalFallback())
	ctx := context.Background()
	accountKey := "integration-account-key"
	failKey := loginFailKeyPrefix + accountKey
	blockKey := loginBlockKeyPrefix + accountKey

	if repo.IsBlocked(ctx, accountKey) {
		t.Fatal("初始不应处于暂停状态")
	}
	var count int64
	for i := 0; i < 5; i++ {
		count = repo.RecordFailure(ctx, accountKey, 5*time.Minute)
	}
	if count != 5 {
		t.Fatalf("累计失败次数 = %d, want 5", count)
	}
	ttl, err := rds.TTL(ctx, failKey).Result()
	if err != nil {
		t.Fatalf("读取失败计数 TTL 失败: %v", err)
	}
	if ttl <= 0 || ttl > 5*time.Minute {
		t.Fatalf("失败计数 TTL = %v, want (0, 5m]", ttl)
	}

	repo.Block(ctx, accountKey, 5*time.Minute)
	if !repo.IsBlocked(ctx, accountKey) {
		t.Fatal("达到阈值必须暂停登录")
	}
	ttl, err = rds.TTL(ctx, blockKey).Result()
	if err != nil {
		t.Fatalf("读取暂停键 TTL 失败: %v", err)
	}
	if ttl <= 0 || ttl > 5*time.Minute {
		t.Fatalf("暂停键 TTL = %v, want (0, 5m]（到期自动解除）", ttl)
	}

	repo.Reset(ctx, accountKey)
	if repo.IsBlocked(ctx, accountKey) {
		t.Fatal("登录成功后必须解除暂停")
	}
	if n, _ := rds.Exists(ctx, failKey, blockKey).Result(); n != 0 {
		t.Fatalf("登录成功后残留键 = %d, want 0（计数清零）", n)
	}
}

// TestLoginThrottleRedisBlockExpires ⑤ 暂停到期自动解除（短 TTL 复现，生产为 5 分钟）
func TestLoginThrottleRedisBlockExpires(t *testing.T) {
	rds := newTestRedisClient(t)
	repo := NewLoginThrottleRepo(rds, NewLocalFallback())
	ctx := context.Background()
	accountKey := "integration-expiring-key"

	repo.Block(ctx, accountKey, 80*time.Millisecond)
	if !repo.IsBlocked(ctx, accountKey) {
		t.Fatal("暂停期内必须拒绝登录")
	}
	time.Sleep(120 * time.Millisecond)
	if repo.IsBlocked(ctx, accountKey) {
		t.Fatal("TTL 到期后必须自动解除，无需人工")
	}
}

// TestLoginThrottleRedisFailureWindowExpires 失败计数窗口到期后重新计数
func TestLoginThrottleRedisFailureWindowExpires(t *testing.T) {
	rds := newTestRedisClient(t)
	repo := NewLoginThrottleRepo(rds, NewLocalFallback())
	ctx := context.Background()
	accountKey := "integration-window-key"

	repo.RecordFailure(ctx, accountKey, 80*time.Millisecond)
	if got := repo.RecordFailure(ctx, accountKey, 80*time.Millisecond); got != 2 {
		t.Fatalf("窗口内累计 = %d, want 2", got)
	}
	time.Sleep(120 * time.Millisecond)
	if got := repo.RecordFailure(ctx, accountKey, 80*time.Millisecond); got != 1 {
		t.Fatalf("窗口到期后重新计数 = %d, want 1", got)
	}
}
