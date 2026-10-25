// 会话清理用例回归（nao-todo-server#56 / DEF-14 / AC7）：
// 过期 8d 回收、过期 6d 保留、活跃 0 删除、已登出按独立宽限期回收、
// 批数上限共享预算、多副本重复执行幂等。
package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"naotodoserver/application/auth/dto"
	"naotodoserver/domain/identity/repositories"
)

// cleanupSessionRow 内存种子行（模拟 user_sessions 一行）
type cleanupSessionRow struct {
	label     string
	expiredAt time.Time
	deletedAt *time.Time
}

// fakeCleanupSessionRepo 只实现清理用到的两个仓储方法；
// 其余接口方法由内嵌 nil 占位（清理路径不会调用）。
// 删除语义按仓储 raw SQL 的谓词复刻：expired_at < cutoff 且 deleted_at IS NULL；
// deleted_at 非空且 < cutoff；均按 limit 截断 ⇒ 可用于断言宽限期与批量上限。
type fakeCleanupSessionRepo struct {
	repositories.UserSession
	mu           sync.Mutex
	rows         []*cleanupSessionRow
	expiredCalls int
	revokedCalls int
}

func (f *fakeCleanupSessionRepo) DeleteExpiredBefore(
	_ context.Context,
	cutoff time.Time,
	limit int,
) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.expiredCalls++
	return f.deleteLocked(limit, func(r *cleanupSessionRow) bool {
		return r.deletedAt == nil && r.expiredAt.Before(cutoff)
	}), nil
}

func (f *fakeCleanupSessionRepo) DeleteRevokedBefore(
	_ context.Context,
	cutoff time.Time,
	limit int,
) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.revokedCalls++
	return f.deleteLocked(limit, func(r *cleanupSessionRow) bool {
		return r.deletedAt != nil && r.deletedAt.Before(cutoff)
	}), nil
}

// deleteLocked 删除前 limit 条匹配行，返回删除行数
func (f *fakeCleanupSessionRepo) deleteLocked(
	limit int,
	match func(r *cleanupSessionRow) bool,
) int64 {
	var kept []*cleanupSessionRow
	var removed int64
	for _, row := range f.rows {
		if removed < int64(limit) && match(row) {
			removed++
			continue
		}
		kept = append(kept, row)
	}
	f.rows = kept
	return removed
}

// tokens 返回当前存活行的种子标识集合
func (f *fakeCleanupSessionRepo) tokens() map[string]bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make(map[string]bool, len(f.rows))
	for _, row := range f.rows {
		result[row.label] = true
	}
	return result
}

func newCleanupTestApp(repo repositories.UserSession) AuthApp {
	// 清理路径只用 sessionRepo；identityDomain / userRepo 不参与
	return NewAuthApp(nil, nil, repo)
}

func cleanupOptions(now time.Time) dto.CleanupSessionsOptions {
	return dto.CleanupSessionsOptions{
		ExpiredBefore: now.AddDate(0, 0, -7),
		RevokedBefore: now.Add(-24 * time.Hour),
		BatchSize:     500,
		MaxBatches:    20,
		BatchPause:    0,
	}
}

// TestCleanupExpiredSessions_OnlyBeyondGrace 只回收越过宽限期的行，活跃行删除数必须为 0
func TestCleanupExpiredSessions_OnlyBeyondGrace(t *testing.T) {
	now := time.Date(2026, 10, 25, 3, 10, 0, 0, time.UTC)
	twoDaysAgo := now.Add(-48 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	repo := &fakeCleanupSessionRepo{rows: []*cleanupSessionRow{
		// 过期 8d（越过 7d 宽限）⇒ 回收
		{label: "expired-8d", expiredAt: now.AddDate(0, 0, -8)},
		// 过期 6d（宽限内）⇒ 保留
		{label: "expired-6d", expiredAt: now.AddDate(0, 0, -6)},
		// 活跃（+1d）⇒ 保留，且删除数必须为 0
		{label: "active", expiredAt: now.AddDate(0, 0, 1)},
		// 登出 2d（越过 24h 宽限）⇒ 回收
		{label: "revoked-2d", expiredAt: twoDaysAgo, deletedAt: &twoDaysAgo},
		// 登出 2h（宽限内）⇒ 保留
		{label: "revoked-2h", expiredAt: twoHoursAgo, deletedAt: &twoHoursAgo},
		// 过期 8d 但 2h 前才登出 ⇒ 两条宽限期独立，登出宽限未到 ⇒ 保留
		{label: "expired-8d-revoked-2h", expiredAt: now.AddDate(0, 0, -8), deletedAt: &twoHoursAgo},
	}}

	removedExpired, removedRevoked, err := newCleanupTestApp(repo).
		CleanupExpiredSessions(context.Background(), cleanupOptions(now))
	if err != nil {
		t.Fatalf("CleanupExpiredSessions: %v", err)
	}
	if removedExpired != 1 {
		t.Fatalf("removedExpired = %d, want 1", removedExpired)
	}
	if removedRevoked != 1 {
		t.Fatalf("removedRevoked = %d, want 1", removedRevoked)
	}

	alive := repo.tokens()
	// 活跃行（active）必须仍在：等价于「活跃会话删除数 = 0」
	for _, kept := range []string{"expired-6d", "active", "revoked-2h", "expired-8d-revoked-2h"} {
		if !alive[kept] {
			t.Fatalf("%s 不应被删除", kept)
		}
	}
	for _, gone := range []string{"expired-8d", "revoked-2d"} {
		if alive[gone] {
			t.Fatalf("%s 应被回收", gone)
		}
	}
}

// TestCleanupExpiredSessions_BatchBudget 批量上限：过期 + 已登出共享 maxBatches 预算
func TestCleanupExpiredSessions_BatchBudget(t *testing.T) {
	now := time.Date(2026, 10, 25, 3, 10, 0, 0, time.UTC)
	twoDaysAgo := now.Add(-48 * time.Hour)
	repo := &fakeCleanupSessionRepo{}
	for i := 0; i < 61; i++ {
		repo.rows = append(repo.rows, &cleanupSessionRow{
			label:     "expired-" + time.Duration(i).String(),
			expiredAt: now.AddDate(0, 0, -8),
		})
	}
	repo.rows = append(repo.rows, &cleanupSessionRow{
		label: "revoked-2d", expiredAt: twoDaysAgo, deletedAt: &twoDaysAgo,
	})

	opts := cleanupOptions(now)
	opts.BatchSize = 3
	opts.MaxBatches = 20
	removedExpired, removedRevoked, err := newCleanupTestApp(repo).
		CleanupExpiredSessions(context.Background(), opts)
	if err != nil {
		t.Fatalf("CleanupExpiredSessions: %v", err)
	}
	if removedExpired != 60 {
		t.Fatalf("removedExpired = %d, want 60（3×20 批预算）", removedExpired)
	}
	if removedRevoked != 0 {
		t.Fatalf("removedRevoked = %d, want 0（批数预算已被过期阶段耗尽）", removedRevoked)
	}
	if repo.expiredCalls != 20 {
		t.Fatalf("过期阶段删除调用 = %d, want 20", repo.expiredCalls)
	}
	if repo.revokedCalls != 0 {
		t.Fatalf("已登出阶段删除调用 = %d, want 0", repo.revokedCalls)
	}
	if len(repo.tokens()) != 2 {
		t.Fatalf("剩余行 = %d, want 2（1 条过期 + 1 条登出）", len(repo.tokens()))
	}
}

// TestCleanupExpiredSessions_Idempotent 多副本重复执行（含并发）幂等：第二次删除 0 行
func TestCleanupExpiredSessions_Idempotent(t *testing.T) {
	now := time.Date(2026, 10, 25, 3, 10, 0, 0, time.UTC)
	twoDaysAgo := now.Add(-48 * time.Hour)
	repo := &fakeCleanupSessionRepo{}
	for i := 0; i < 3; i++ {
		repo.rows = append(repo.rows, &cleanupSessionRow{
			label:     "expired-" + time.Duration(i).String(),
			expiredAt: now.AddDate(0, 0, -8),
		})
		repo.rows = append(repo.rows, &cleanupSessionRow{
			label:     "revoked-" + time.Duration(i).String(),
			expiredAt: twoDaysAgo,
			deletedAt: &twoDaysAgo,
		})
		repo.rows = append(repo.rows, &cleanupSessionRow{
			label:     "active-" + time.Duration(i).String(),
			expiredAt: now.AddDate(0, 0, 1),
		})
	}
	app := newCleanupTestApp(repo)
	opts := cleanupOptions(now)

	// 两个副本同时执行：谓词删除，不重复计数
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			if _, _, err := app.CleanupExpiredSessions(context.Background(), opts); err != nil {
				t.Errorf("并发 CleanupExpiredSessions: %v", err)
			}
		}()
	}
	wg.Wait()

	if alive := repo.tokens(); len(alive) != 3 {
		t.Fatalf("剩余行 = %d, want 3（仅活跃行）", len(alive))
	}
	// 再跑一次：无行可删 ⇒ 幂等
	removedExpired, removedRevoked, err := app.CleanupExpiredSessions(context.Background(), opts)
	if err != nil {
		t.Fatalf("重复 CleanupExpiredSessions: %v", err)
	}
	if removedExpired != 0 || removedRevoked != 0 {
		t.Fatalf("重复执行删除 = (%d, %d), want (0, 0)", removedExpired, removedRevoked)
	}
}
