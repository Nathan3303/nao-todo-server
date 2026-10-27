package identity

import (
	"context"
	"testing"
	"time"
)

// TestLoginThrottleFallbackWhenRedisDown Redis 故障 ⇒ 账号节流退化为进程内计数，
// 仍能在 5 次失败后暂停登录（与 SEC-10 共用本地兜底）。
func TestLoginThrottleFallbackWhenRedisDown(t *testing.T) {
	logs := captureLogs(t)
	repo := &LoginThrottleRepoImpl{rds: newDownRedisClient(t), local: NewLocalFallback()}
	ctx := context.Background()
	accountKey := "0123456789abcdef"

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
	repo.Block(ctx, accountKey, 5*time.Minute)
	if !repo.IsBlocked(ctx, accountKey) {
		t.Fatal("降级路径下达到阈值必须暂停登录")
	}
	repo.Reset(ctx, accountKey)
	if repo.IsBlocked(ctx, accountKey) {
		t.Fatal("Reset 后必须解除暂停")
	}
	if got := repo.RecordFailure(ctx, accountKey, 5*time.Minute); got != 1 {
		t.Fatalf("Reset 后重新计数 = %d, want 1", got)
	}

	if findLogEntry(logEntries(t, logs), "login.throttle.degraded") == nil {
		t.Fatal("降级必须记录 login.throttle.degraded 结构化日志")
	}
}

// TestLoginThrottleFallbackExpires 本地兜底的暂停状态到期自动解除
func TestLoginThrottleFallbackExpires(t *testing.T) {
	captureLogs(t)
	repo := &LoginThrottleRepoImpl{rds: newDownRedisClient(t), local: NewLocalFallback()}
	ctx := context.Background()
	accountKey := "fedcba9876543210"

	repo.Block(ctx, accountKey, 40*time.Millisecond)
	if !repo.IsBlocked(ctx, accountKey) {
		t.Fatal("暂停期内应拒绝登录")
	}
	time.Sleep(60 * time.Millisecond)
	if repo.IsBlocked(ctx, accountKey) {
		t.Fatal("到期后必须自动解除")
	}
}

// TestLoginThrottleFallbackWindowReset 本地失败计数窗口到期后重新计数
func TestLoginThrottleFallbackWindowReset(t *testing.T) {
	captureLogs(t)
	repo := &LoginThrottleRepoImpl{rds: newDownRedisClient(t), local: NewLocalFallback()}
	ctx := context.Background()
	accountKey := "0011223344556677"

	repo.RecordFailure(ctx, accountKey, 40*time.Millisecond)
	if got := repo.RecordFailure(ctx, accountKey, 40*time.Millisecond); got != 2 {
		t.Fatalf("窗口内累计 = %d, want 2", got)
	}
	time.Sleep(60 * time.Millisecond)
	if got := repo.RecordFailure(ctx, accountKey, 40*time.Millisecond); got != 1 {
		t.Fatalf("窗口到期后重新计数 = %d, want 1", got)
	}
}
