//go:build integration

// T466 集成测试（真实 MySQL）：/sync/push 清单 deactivedAt 三态行级落地（DEF-46）。
//
// 全链路：HTTP JSON 载荷 → gin 绑定 → 控制器 → 真实应用层 → 真实仓储 → MySQL 列断言。
// 复用同包（build tag integration）的 TestMain / newProjectSyncController /
// doNullClearPush / projectPushBody 等公共装置。
// 运行：TZ=UTC go test -tags integration -count=1 -run TestSyncProjectDeactived -v ./interfaces/controllers/
package controllers

import (
	"database/sql"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/interfaces/types"
)

// seedProjectRowDeactived 灌入清单行（deactivedAt 非零 ⇒ 停用态），返回库中快照。
func seedProjectRowDeactived(t *testing.T, uid, pid int64, name string, deactivedAt *time.Time) models.Project {
	t.Helper()
	base := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	p := &models.Project{
		ModelBase: models.ModelBase{ID: pid, CreatedAt: base, UpdatedAt: base},
		UserId:    uid,
		Name:      name,
	}
	if deactivedAt != nil {
		p.DeactivedAt = sql.NullTime{Time: *deactivedAt, Valid: true}
	}
	if err := nullClearDB.Create(p).Error; err != nil {
		t.Fatalf("seed 清单 %d 失败: %v", pid, err)
	}
	return readProjectRow(t, pid)
}

// TestSyncProjectDeactivedAtNullClears ① DEF-46 主场景：清单恢复（`deactivedAt: null`）⇒
// `deactived_at` 变 NULL，且回执 applied。
func TestSyncProjectDeactivedAtNullClears(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9901), int64(9902)
	cleanProjectNullClear(t, uid)
	deactivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRowDeactived(t, uid, pid, "T466 清单", &deactivedAt)
	if !seed.DeactivedAt.Valid {
		t.Fatalf("前置失败：seed 清单应处于停用态")
	}

	body := projectPushBody(pid, uid, "T466 清单", idutil.FormatTimeMilli(seed.UpdatedAt), `"deactivedAt":null`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readProjectRow(t, pid)
	if after.DeactivedAt.Valid {
		t.Fatalf("修复未生效：deactived_at 仍为 %v（应被 null 清空）",
			after.DeactivedAt.Time.UTC().Format(time.RFC3339))
	}
	if !after.UpdatedAt.After(seed.UpdatedAt) {
		t.Fatalf("覆盖分支应 bump updated_at: %v → %v", seed.UpdatedAt, after.UpdatedAt)
	}
	t.Logf("① deactivedAt:null ⇒ outcome=%s · deactived_at %v → NULL",
		results[0].Outcome, deactivedAt.UTC().Format(time.RFC3339))
}

// TestSyncProjectDeactivedAtAbsentKeeps ② 真三态回归锁：缺省（旧客户端）⇒ 不清空，其余字段照常覆盖。
func TestSyncProjectDeactivedAtAbsentKeeps(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9911), int64(9912)
	cleanProjectNullClear(t, uid)
	deactivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRowDeactived(t, uid, pid, "T466 清单", &deactivedAt)

	body := projectPushBody(pid, uid, "T466 改名", idutil.FormatTimeMilli(seed.UpdatedAt), "")
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readProjectRow(t, pid)
	if !after.DeactivedAt.Valid || !after.DeactivedAt.Time.Equal(seed.DeactivedAt.Time) {
		t.Fatalf("absent 不得改 deactived_at: %v → %+v", seed.DeactivedAt.Time, after.DeactivedAt)
	}
	if after.Name != "T466 改名" {
		t.Fatalf("其它字段应照常覆盖: name = %q", after.Name)
	}
	t.Logf("② absent ⇒ deactived_at 保持 %v · name → %s",
		after.DeactivedAt.Time.UTC().Format(time.RFC3339), after.Name)
}

// TestSyncProjectDeactivedAtValueWrites ③ 停用：推送具体时间 ⇒ 写入该时间。
func TestSyncProjectDeactivedAtValueWrites(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9921), int64(9922)
	cleanProjectNullClear(t, uid)
	seed := seedProjectRowDeactived(t, uid, pid, "T466 清单", nil)
	if seed.DeactivedAt.Valid {
		t.Fatalf("前置失败：seed 清单应为未停用态")
	}

	wantAt := time.Now().UTC().Truncate(time.Second).Add(-5 * time.Minute)
	body := projectPushBody(pid, uid, "T466 清单", idutil.FormatTimeMilli(seed.UpdatedAt),
		`"deactivedAt":"`+wantAt.Format(time.RFC3339)+`"`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readProjectRow(t, pid)
	if !after.DeactivedAt.Valid || !after.DeactivedAt.Time.Equal(wantAt) {
		t.Fatalf("deactived_at = %+v, want %v", after.DeactivedAt, wantAt)
	}
	t.Logf("③ deactivedAt=%s ⇒ deactived_at 写入 %v", wantAt.Format(time.RFC3339),
		after.DeactivedAt.Time.UTC().Format(time.RFC3339))
}
