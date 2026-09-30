//go:build integration

// T466 集成测试（真实 MySQL）：/sync/push 常用番茄工作 archivedAt 三态行级落地（DEF-43）。
//
// 全链路：HTTP JSON 载荷 → gin 绑定 → 控制器 → 真实应用层 → 真实仓储 → MySQL 列断言。
// 复用同包（build tag integration）的 TestMain / doNullClearPush。
// 运行：TZ=UTC go test -tags integration -count=1 -run TestSyncPomodoroArchived -v ./interfaces/controllers/
package controllers

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroService "naotodoserver/domain/pomodoro/service"
	"naotodoserver/infrastructure/persistence/models"
	pomodoroPersistence "naotodoserver/infrastructure/persistence/pomodoro"
	"naotodoserver/interfaces/types"
)

// newPomodoroSyncController 真实装配：番茄仓储（MySQL）→ 领域 → 应用 → sync 控制器。
// 其余表应用传 nil：本用例载荷不含那些表，Push 不会触碰。
func newPomodoroSyncController() *SyncController {
	pomodoroRepoInst := pomodoroPersistence.NewPomodoroRepo(nullClearDB)
	recordRepoInst := pomodoroPersistence.NewPomodoroRecordRepo(nullClearDB)
	domain := pomodoroService.NewPomodoroDomain(recordRepoInst, pomodoroRepoInst)
	app := pomodoroApp.NewPomodoroApp(domain, recordRepoInst, pomodoroRepoInst)
	return &SyncController{pomodoroApp: app}
}

func cleanPomodoroNullClear(t *testing.T, uid int64) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM pomodoro_records WHERE user_id = ?",
		"DELETE FROM pomodoros WHERE user_id = ?",
	} {
		if err := nullClearDB.Exec(stmt, uid).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// seedPomodoroRow 灌入常用番茄工作行（archivedAt 非零 ⇒ 归档态），返回库中快照。
func seedPomodoroRow(t *testing.T, uid, pid int64, name string, archivedAt *time.Time) models.Pomodoro {
	t.Helper()
	base := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	p := &models.Pomodoro{
		ModelBase: models.ModelBase{ID: pid, CreatedAt: base, UpdatedAt: base},
		UserId:    uid,
		Type:      1,
		Name:      name,
		Duration:  1500,
	}
	if archivedAt != nil {
		p.ArchivedAt = sql.NullTime{Time: *archivedAt, Valid: true}
	}
	if err := nullClearDB.Create(p).Error; err != nil {
		t.Fatalf("seed 番茄 %d 失败: %v", pid, err)
	}
	return readPomodoroRow(t, pid)
}

func readPomodoroRow(t *testing.T, pid int64) models.Pomodoro {
	t.Helper()
	var m models.Pomodoro
	if err := nullClearDB.Unscoped().First(&m, "id = ?", pid).Error; err != nil {
		t.Fatalf("读取番茄 %d 失败: %v", pid, err)
	}
	return m
}

// pomodoroPushBody 客户端同形番茄推送载荷；archived 为附加字段 JSON 片段（可空）。
func pomodoroPushBody(pid int64, name, base string, archived string) string {
	body := fmt.Sprintf(`{"pomodoros":[{"id":"%d","type":1,"name":"%s","duration":1500,"baseUpdatedAt":"%s"`,
		pid, name, base)
	if archived != "" {
		body += "," + archived
	}
	return body + `}]}`
}

// TestSyncPomodoroArchivedAtNullClears ① DEF-43 主场景：取消归档（`archivedAt: null`）⇒
// `archived_at` 变 NULL，且回执 applied。
func TestSyncPomodoroArchivedAtNullClears(t *testing.T) {
	ctrl := newPomodoroSyncController()
	const uid, pid = int64(9951), int64(9952)
	cleanPomodoroNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedPomodoroRow(t, uid, pid, "T466 番茄", &archivedAt)
	if !seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 番茄应处于归档态")
	}

	body := pomodoroPushBody(pid, "T466 番茄", idutil.FormatTimeMilli(seed.UpdatedAt), `"archivedAt":null`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readPomodoroRow(t, pid)
	if after.ArchivedAt.Valid {
		t.Fatalf("修复未生效：archived_at 仍为 %v（应被 null 清空）",
			after.ArchivedAt.Time.UTC().Format(time.RFC3339))
	}
	t.Logf("① archivedAt:null ⇒ outcome=%s · archived_at %v → NULL",
		results[0].Outcome, archivedAt.UTC().Format(time.RFC3339))
}

// TestSyncPomodoroArchivedAtAbsentKeeps ② 真三态回归锁：缺省（旧客户端）⇒ 不清空。
func TestSyncPomodoroArchivedAtAbsentKeeps(t *testing.T) {
	ctrl := newPomodoroSyncController()
	const uid, pid = int64(9961), int64(9962)
	cleanPomodoroNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedPomodoroRow(t, uid, pid, "T466 番茄", &archivedAt)

	body := pomodoroPushBody(pid, "T466 改名", idutil.FormatTimeMilli(seed.UpdatedAt), "")
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readPomodoroRow(t, pid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(seed.ArchivedAt.Time) {
		t.Fatalf("absent 不得改 archived_at: %v → %+v", seed.ArchivedAt.Time, after.ArchivedAt)
	}
	if after.Name != "T466 改名" {
		t.Fatalf("其它字段应照常覆盖: name = %q", after.Name)
	}
	t.Logf("② absent ⇒ archived_at 保持 %v · name → %s",
		after.ArchivedAt.Time.UTC().Format(time.RFC3339), after.Name)
}

// TestSyncPomodoroArchivedAtValueWrites ③ 归档：推送具体时间 ⇒ 写入该时间。
func TestSyncPomodoroArchivedAtValueWrites(t *testing.T) {
	ctrl := newPomodoroSyncController()
	const uid, pid = int64(9971), int64(9972)
	cleanPomodoroNullClear(t, uid)
	seed := seedPomodoroRow(t, uid, pid, "T466 番茄", nil)
	if seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 番茄应为未归档态")
	}

	wantAt := time.Now().UTC().Truncate(time.Second).Add(-5 * time.Minute)
	body := pomodoroPushBody(pid, "T466 番茄", idutil.FormatTimeMilli(seed.UpdatedAt),
		`"archivedAt":"`+wantAt.Format(time.RFC3339)+`"`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readPomodoroRow(t, pid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(wantAt) {
		t.Fatalf("archived_at = %+v, want %v", after.ArchivedAt, wantAt)
	}
	t.Logf("③ archivedAt=%s ⇒ archived_at 写入 %v", wantAt.Format(time.RFC3339),
		after.ArchivedAt.Time.UTC().Format(time.RFC3339))
}
