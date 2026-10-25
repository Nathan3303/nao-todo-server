// 会话清理配置回归（nao-todo-server#56）：SESSION_CLEANUP_* 环境变量覆盖 + 缺省兜底。
package conf

import "testing"

var sessionCleanupEnvKeys = []string{
	"SESSION_CLEANUP_ENABLED",
	"SESSION_CLEANUP_CRON",
	"SESSION_CLEANUP_RETENTION_DAYS",
	"SESSION_CLEANUP_REVOKED_RETENTION_HOURS",
	"SESSION_CLEANUP_BATCH_SIZE",
	"SESSION_CLEANUP_MAX_BATCHES",
	"SESSION_CLEANUP_RUN_ON_START",
}

// clearSessionCleanupEnv 清空相关环境变量，保证用例互不影响
func clearSessionCleanupEnv(t *testing.T) {
	t.Helper()
	for _, key := range sessionCleanupEnvKeys {
		t.Setenv(key, "")
	}
}

// TestOverrideSessionCleanupWithEnv 环境变量覆盖全部四组参数 + 两个附加开关
func TestOverrideSessionCleanupWithEnv(t *testing.T) {
	clearSessionCleanupEnv(t)
	t.Setenv("SESSION_CLEANUP_ENABLED", "true")
	t.Setenv("SESSION_CLEANUP_CRON", "*/5 * * * *")
	t.Setenv("SESSION_CLEANUP_RETENTION_DAYS", "10")
	t.Setenv("SESSION_CLEANUP_REVOKED_RETENTION_HOURS", "48")
	t.Setenv("SESSION_CLEANUP_BATCH_SIZE", "100")
	t.Setenv("SESSION_CLEANUP_MAX_BATCHES", "5")
	t.Setenv("SESSION_CLEANUP_RUN_ON_START", "true")

	sc := &SessionCleanup{}
	overrideSessionCleanupWithEnv(sc)
	sc.applyDefaults()

	if !sc.Enabled || !sc.RunOnStart {
		t.Fatalf("Enabled/RunOnStart = %t/%t, want true/true", sc.Enabled, sc.RunOnStart)
	}
	if sc.Cron != "*/5 * * * *" {
		t.Fatalf("Cron = %q, want %q", sc.Cron, "*/5 * * * *")
	}
	if sc.RetentionDays != 10 || sc.RevokedRetentionHours != 48 {
		t.Fatalf(
			"宽限期 = %d天/%d小时, want 10天/48小时",
			sc.RetentionDays,
			sc.RevokedRetentionHours,
		)
	}
	if sc.BatchSize != 100 || sc.MaxBatches != 5 {
		t.Fatalf("批量 = %d×%d, want 100×5", sc.BatchSize, sc.MaxBatches)
	}
}

// TestSessionCleanupEnvDisables 回滚开关：SESSION_CLEANUP_ENABLED=false 覆盖 yaml 的 true
func TestSessionCleanupEnvDisables(t *testing.T) {
	clearSessionCleanupEnv(t)
	t.Setenv("SESSION_CLEANUP_ENABLED", "false")

	sc := &SessionCleanup{Enabled: true}
	overrideSessionCleanupWithEnv(sc)

	if sc.Enabled {
		t.Fatal("SESSION_CLEANUP_ENABLED=false 应覆盖为停用")
	}
}

// TestSessionCleanupDefaults 缺省配置：批量/周期/宽限期兜底，但 enabled 零值保持停用（宁可不删）
func TestSessionCleanupDefaults(t *testing.T) {
	clearSessionCleanupEnv(t)

	sc := &SessionCleanup{}
	sc.applyDefaults()

	if sc.Cron != "10 3 * * *" {
		t.Fatalf("Cron = %q, want %q", sc.Cron, "10 3 * * *")
	}
	if sc.RetentionDays != 7 || sc.RevokedRetentionHours != 24 {
		t.Fatalf(
			"宽限期 = %d天/%d小时, want 7天/24小时",
			sc.RetentionDays,
			sc.RevokedRetentionHours,
		)
	}
	if sc.BatchSize != 500 || sc.MaxBatches != 20 {
		t.Fatalf("批量 = %d×%d, want 500×20", sc.BatchSize, sc.MaxBatches)
	}
	if sc.Enabled {
		t.Fatal("缺省配置下 enabled 应保持 false（不删除任何行）")
	}
}
