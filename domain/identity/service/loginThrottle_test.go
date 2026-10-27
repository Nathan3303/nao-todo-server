package service

import (
	"context"
	"testing"
	"time"

	"naotodoserver/domain/identity/valueobjects"
)

// fakeLoginThrottle 内存版账号节流仓储：
// 记录调用参数（键/窗口/时长），用于验证领域策略与"不落明文邮箱"。
type fakeLoginThrottle struct {
	counts   map[string]int64
	blocked  map[string]bool
	windows  []time.Duration
	blocks   map[string]time.Duration
	resets   []string
	blockCue []string
}

func newFakeLoginThrottle() *fakeLoginThrottle {
	return &fakeLoginThrottle{
		counts:  make(map[string]int64),
		blocked: make(map[string]bool),
		blocks:  make(map[string]time.Duration),
	}
}

func (f *fakeLoginThrottle) RecordFailure(_ context.Context, accountKey string, window time.Duration) int64 {
	f.counts[accountKey]++
	f.windows = append(f.windows, window)
	return f.counts[accountKey]
}

func (f *fakeLoginThrottle) Block(_ context.Context, accountKey string, duration time.Duration) {
	f.blocked[accountKey] = true
	f.blocks[accountKey] = duration
	f.blockCue = append(f.blockCue, accountKey)
}

func (f *fakeLoginThrottle) IsBlocked(_ context.Context, accountKey string) bool {
	return f.blocked[accountKey]
}

func (f *fakeLoginThrottle) Reset(_ context.Context, accountKey string) {
	delete(f.blocked, accountKey)
	delete(f.counts, accountKey)
	f.resets = append(f.resets, accountKey)
}

// newThrottleDomain 只装配账号节流所需的依赖（其余仓储不参与本用例）
func newThrottleDomain(f *fakeLoginThrottle) *identityDomainImpl {
	return &identityDomainImpl{loginThrottle: f}
}

// TestRecordLoginFailureBlocksAtThreshold 第 5 次失败 ⇒ 触发暂停（5 分钟）；
// 第 4 次之前不暂停（避免误锁正常手误）。
func TestRecordLoginFailureBlocksAtThreshold(t *testing.T) {
	ctx := context.Background()
	email := "victim@example.com"
	accountKey := valueobjects.AccountKey(email)
	f := newFakeLoginThrottle()
	domain := newThrottleDomain(f)

	for i := 1; i < int(LoginFailureThreshold); i++ {
		domain.RecordLoginFailure(ctx, email)
		if f.blocked[accountKey] {
			t.Fatalf("第 %d 次失败不应暂停登录", i)
		}
	}
	domain.RecordLoginFailure(ctx, email)
	if !f.blocked[accountKey] {
		t.Fatalf("第 %d 次失败必须暂停登录", LoginFailureThreshold)
	}
	if got := f.blocks[accountKey]; got != LoginBlockDuration {
		t.Fatalf("暂停时长 = %v, want %v（到期自动解除）", got, LoginBlockDuration)
	}
	if got := f.windows[0]; got != LoginFailureWindow {
		t.Fatalf("失败计数窗口 = %v, want %v", got, LoginFailureWindow)
	}
	// 阈值/窗口/暂停时长即产品口径（5 次 / 5 分钟 / 5 分钟）
	if LoginFailureThreshold != 5 || LoginFailureWindow != 5*time.Minute ||
		LoginBlockDuration != 5*time.Minute {
		t.Fatalf("策略常量被改动: threshold=%d window=%v block=%v",
			LoginFailureThreshold, LoginFailureWindow, LoginBlockDuration)
	}
}

// TestClearLoginFailuresResetsAccount 登录成功 ⇒ 清除计数与暂停状态
func TestClearLoginFailuresResetsAccount(t *testing.T) {
	ctx := context.Background()
	email := "victim@example.com"
	accountKey := valueobjects.AccountKey(email)
	f := newFakeLoginThrottle()
	domain := newThrottleDomain(f)

	for i := int64(0); i < LoginFailureThreshold; i++ {
		domain.RecordLoginFailure(ctx, email)
	}
	if !domain.IsLoginBlocked(ctx, email) {
		t.Fatal("达到阈值后账号应处于暂停状态")
	}

	domain.ClearLoginFailures(ctx, email)
	if domain.IsLoginBlocked(ctx, email) {
		t.Fatal("成功后必须清除暂停状态")
	}
	if f.counts[accountKey] != 0 {
		t.Fatalf("成功后失败计数 = %d, want 0", f.counts[accountKey])
	}
}

// TestLoginThrottleKeysAreHashed 领域层只把哈希标识交给仓储：不落邮箱明文，
// 且大小写不同的同一邮箱共用计数（规范化口径）。
func TestLoginThrottleKeysAreHashed(t *testing.T) {
	ctx := context.Background()
	f := newFakeLoginThrottle()
	domain := newThrottleDomain(f)

	domain.RecordLoginFailure(ctx, "Victim@Example.COM")
	domain.IsLoginBlocked(ctx, "victim@example.com")

	want := valueobjects.AccountKey("victim@example.com")
	if len(f.counts) != 1 || f.counts[want] != 1 {
		t.Fatalf("失败计数键 = %v, want 仅 %q", f.counts, want)
	}
	if _, ok := f.counts["victim@example.com"]; ok {
		t.Fatal("仓储不得收到邮箱明文")
	}
}
