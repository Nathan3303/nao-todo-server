// 过期会话清理任务单测：时间上界按宽限期计算、参数透传、失败不 panic。
package cron

import (
	"context"
	"errors"
	"testing"
	"time"

	authApp "naotodoserver/application/auth"
	"naotodoserver/application/auth/dto"
	"naotodoserver/conf"
)

// fakeCleanupAuthApp 记录清理入参并返回预设行数
type fakeCleanupAuthApp struct {
	authApp.AuthApp
	calls    int
	lastOpts dto.CleanupSessionsOptions
	err      error
}

func (f *fakeCleanupAuthApp) CleanupExpiredSessions(
	_ context.Context,
	opts dto.CleanupSessionsOptions,
) (int64, int64, error) {
	f.calls++
	f.lastOpts = opts
	return 2, 3, f.err
}

func TestCleanExpiredSessionsJob_ComputesCutoffFromConfig(t *testing.T) {
	now := time.Date(2026, 10, 25, 3, 10, 0, 0, time.UTC)
	fakeApp := &fakeCleanupAuthApp{}
	job := NewCleanExpiredSessionsJob(fakeApp, conf.SessionCleanup{
		Enabled:               true,
		RetentionDays:         7,
		RevokedRetentionHours: 24,
		BatchSize:             500,
		MaxBatches:            20,
	})
	job.Now = func() time.Time { return now }
	job.Run()

	if fakeApp.calls != 1 {
		t.Fatalf("清理调用次数 = %d, want 1", fakeApp.calls)
	}
	opts := fakeApp.lastOpts
	if want := now.AddDate(0, 0, -7); !opts.ExpiredBefore.Equal(want) {
		t.Fatalf("ExpiredBefore = %v, want %v", opts.ExpiredBefore, want)
	}
	if want := now.Add(-24 * time.Hour); !opts.RevokedBefore.Equal(want) {
		t.Fatalf("RevokedBefore = %v, want %v", opts.RevokedBefore, want)
	}
	if opts.BatchSize != 500 || opts.MaxBatches != 20 {
		t.Fatalf("批量参数 = (%d, %d), want (500, 20)", opts.BatchSize, opts.MaxBatches)
	}
	if opts.BatchPause != 50*time.Millisecond {
		t.Fatalf("BatchPause = %v, want 50ms", opts.BatchPause)
	}
}

func TestCleanExpiredSessionsJob_ErrorDoesNotPanic(t *testing.T) {
	fakeApp := &fakeCleanupAuthApp{err: errors.New("db down")}
	job := NewCleanExpiredSessionsJob(fakeApp, conf.SessionCleanup{
		RetentionDays:         7,
		RevokedRetentionHours: 24,
		BatchSize:             500,
		MaxBatches:            20,
	})
	job.Run()
	if fakeApp.calls != 1 {
		t.Fatalf("清理调用次数 = %d, want 1", fakeApp.calls)
	}
}
