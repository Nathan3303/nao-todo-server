// 会话清理 cron 门控回归（nao-todo-server#56）：
// 开关关闭 ⇒ 不注册、不执行（含 runOnStart）；开启 ⇒ 按配置周期注册，runOnStart 立即跑一次。
package infrastructure

import (
	"context"
	"sync"
	"testing"
	"time"

	"naotodoserver/application"
	authApp "naotodoserver/application/auth"
	"naotodoserver/application/auth/dto"
	"naotodoserver/conf"
	"naotodoserver/infrastructure/cron"

	robfigcron "github.com/robfig/cron/v3"
)

// fakeCronService 记录注册的任务，不实际调度
type fakeCronService struct {
	jobs map[string]robfigcron.Job
}

func newFakeCronService() *fakeCronService {
	return &fakeCronService{jobs: map[string]robfigcron.Job{}}
}

func (f *fakeCronService) Start()                           {}
func (f *fakeCronService) Stop()                            {}
func (f *fakeCronService) AddFunc(_ string, _ func()) error { return nil }
func (f *fakeCronService) AddJob(
	flag string,
	job robfigcron.Job,
) (robfigcron.EntryID, error) {
	f.jobs[flag] = job
	return 0, nil
}

// fakeCleanupAuthApp 记录清理调用（并发安全，runOnStart 走 goroutine）
type fakeCleanupAuthApp struct {
	authApp.AuthApp
	mu   sync.Mutex
	opts []dto.CleanupSessionsOptions
	done chan dto.CleanupSessionsOptions
}

func (f *fakeCleanupAuthApp) CleanupExpiredSessions(
	_ context.Context,
	opts dto.CleanupSessionsOptions,
) (int64, int64, error) {
	f.mu.Lock()
	f.opts = append(f.opts, opts)
	f.mu.Unlock()
	if f.done != nil {
		select {
		case f.done <- opts:
		default:
		}
	}
	return 0, 0, nil
}

func (f *fakeCleanupAuthApp) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.opts)
}

func cleanupTestConfig(enabled, runOnStart bool) *conf.Config {
	return &conf.Config{
		Session: conf.Session{Cleanup: conf.SessionCleanup{
			Enabled:               enabled,
			Cron:                  "10 3 * * *",
			RetentionDays:         7,
			RevokedRetentionHours: 24,
			BatchSize:             500,
			MaxBatches:            20,
			RunOnStart:            runOnStart,
		}},
	}
}

// TestRegisterJobs_SessionCleanupDisabled 开关关闭 ⇒ 不注册、不执行（runOnStart 也不破例）
func TestRegisterJobs_SessionCleanupDisabled(t *testing.T) {
	cronService := newFakeCronService()
	authFake := &fakeCleanupAuthApp{}

	registerJobs(cronService, &application.Services{Auth: authFake}, cleanupTestConfig(false, true))

	if len(cronService.jobs) != 2 {
		t.Fatalf("注册任务数 = %d, want 2（仅既有任务）", len(cronService.jobs))
	}
	if _, ok := cronService.jobs["10 3 * * *"]; ok {
		t.Fatal("开关关闭时不应注册会话清理任务")
	}
	if authFake.calls() != 0 {
		t.Fatalf("开关关闭时清理调用次数 = %d, want 0", authFake.calls())
	}
}

// TestRegisterJobs_SessionCleanupEnabled 开关开启 ⇒ 按配置周期注册；runOnStart=false 不立即执行
func TestRegisterJobs_SessionCleanupEnabled(t *testing.T) {
	cronService := newFakeCronService()
	authFake := &fakeCleanupAuthApp{}

	registerJobs(cronService, &application.Services{Auth: authFake}, cleanupTestConfig(true, false))

	job, ok := cronService.jobs["10 3 * * *"]
	if !ok {
		t.Fatal("开关开启时应按配置周期注册会话清理任务")
	}
	if _, isCleanup := job.(*cron.CleanExpiredSessionsJob); !isCleanup {
		t.Fatalf("注册任务类型 = %T, want *cron.CleanExpiredSessionsJob", job)
	}
	if authFake.calls() != 0 {
		t.Fatalf("runOnStart=false 时清理调用次数 = %d, want 0", authFake.calls())
	}
}

// TestRegisterJobs_SessionCleanupRunOnStart runOnStart=true ⇒ 启动后立即执行一次（部署验证用）
func TestRegisterJobs_SessionCleanupRunOnStart(t *testing.T) {
	cronService := newFakeCronService()
	authFake := &fakeCleanupAuthApp{done: make(chan dto.CleanupSessionsOptions, 1)}

	before := time.Now()
	registerJobs(cronService, &application.Services{Auth: authFake}, cleanupTestConfig(true, true))

	var opts dto.CleanupSessionsOptions
	select {
	case opts = <-authFake.done:
	case <-time.After(2 * time.Second):
		t.Fatal("runOnStart=true 但启动后未执行清理")
	}
	// 上界按执行时刻计算（近似断言，容忍调度延迟）
	if delta := opts.ExpiredBefore.Sub(before.AddDate(0, 0, -7)); delta < -5*time.Second ||
		delta > 5*time.Second {
		t.Fatalf("ExpiredBefore 偏差 = %v, want 约 now-7d", delta)
	}
	if delta := opts.RevokedBefore.Sub(before.Add(-24 * time.Hour)); delta < -5*time.Second ||
		delta > 5*time.Second {
		t.Fatalf("RevokedBefore 偏差 = %v, want 约 now-24h", delta)
	}
	if opts.BatchSize != 500 || opts.MaxBatches != 20 {
		t.Fatalf("批量参数 = (%d, %d), want (500, 20)", opts.BatchSize, opts.MaxBatches)
	}
}
