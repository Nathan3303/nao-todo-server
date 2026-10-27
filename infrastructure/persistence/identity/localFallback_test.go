package identity

import (
	"testing"
	"time"
)

// TestLocalFallbackFixedWindow 固定窗口计数：窗口内累加，到期后重新从 1 开始
func TestLocalFallbackFixedWindow(t *testing.T) {
	local := NewLocalFallback()
	for want := int64(1); want <= 3; want++ {
		if got := local.Incr("k", 40*time.Millisecond); got != want {
			t.Fatalf("第 %d 次计数 = %d", want, got)
		}
	}
	if got := local.Count("k"); got != 3 {
		t.Fatalf("Count = %d, want 3", got)
	}
	time.Sleep(60 * time.Millisecond)
	if got := local.Count("k"); got != 0 {
		t.Fatalf("窗口到期后 Count = %d, want 0（自动失效）", got)
	}
	if got := local.Incr("k", time.Minute); got != 1 {
		t.Fatalf("到期后重新计数 = %d, want 1", got)
	}
}

// TestLocalFallbackReset 登录成功后清零
func TestLocalFallbackReset(t *testing.T) {
	local := NewLocalFallback()
	local.Incr("k", time.Minute)
	local.Incr("k", time.Minute)
	local.Reset("k")
	if got := local.Count("k"); got != 0 {
		t.Fatalf("Reset 后 Count = %d, want 0", got)
	}
}

// TestLocalFallbackSweepsExpiredKeys 过期键被周期清理：键空间有界，不随请求无界增长
func TestLocalFallbackSweepsExpiredKeys(t *testing.T) {
	local := NewLocalFallback()
	for i := 0; i < 200; i++ {
		local.Incr(string(rune('a'+i%26))+string(rune('a'+i/26)), 20*time.Millisecond)
	}
	time.Sleep(40 * time.Millisecond)
	// 触发清理：距上次清理超过周期（测试内直接把 lastSweep 前移）
	local.mu.Lock()
	local.lastSweep = time.Now().Add(-2 * localSweepInterval)
	local.mu.Unlock()
	local.Incr("fresh", time.Minute)

	local.mu.Lock()
	size := len(local.counters)
	local.mu.Unlock()
	if size != 1 {
		t.Fatalf("清理后键数量 = %d, want 1（仅保留未过期键）", size)
	}
}
