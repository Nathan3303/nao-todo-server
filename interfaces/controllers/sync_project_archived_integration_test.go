//go:build integration

// T322 集成测试（真实 MySQL）：/sync/push 清单 archivedAt 三态行级落地（DEF-42）。
//
// 全链路：HTTP JSON 载荷 → gin 绑定 → 控制器 → 真实应用层 → 真实仓储 → MySQL 列断言
//   - 返回体 outcome 断言；并锁定协议约定「服务端不级联」。
//
// 复用同包集成测试的 TestMain / nullClearDB（一个包只允许一个 TestMain，
// 见 interfaces/controllers/sync_null_clear_integration_test.go）。
// 运行：TZ=UTC go test -tags integration -count=1 -run TestSyncProjectArchived -v ./interfaces/controllers/
package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	projectService "naotodoserver/domain/project/service"
	taskValueObjects "naotodoserver/domain/task/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	projectPersistence "naotodoserver/infrastructure/persistence/project"
	taskPersistence "naotodoserver/infrastructure/persistence/task"
	"naotodoserver/interfaces/types"

	"github.com/go-redis/redis/v8"
)

// newProjectSyncController 真实装配：清单仓储（MySQL）→ 领域 → 应用 → sync 控制器。
// 缓存后端指向不可达 Redis（cache 静默降级）；countPublisher 传 nil（发布器内置 nil 保护）。
func newProjectSyncController() *SyncController {
	taskRepoInst := taskPersistence.NewTaskRepo(nullClearDB)
	projectRepoInst := projectPersistence.NewProjectRepo(
		nullClearDB,
		cache.NewCache(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})),
	)
	prefRepo := projectPersistence.NewProjectPreferenceRepo(nullClearDB)
	domain := projectService.NewProjectDomain(projectRepoInst, prefRepo)
	app := projectApp.NewProjectApp(
		domain, dbs.NewTxManager(nullClearDB), projectRepoInst, prefRepo, taskRepoInst, nil,
	)
	return &SyncController{projectApp: app}
}

func cleanProjectNullClear(t *testing.T, uid int64) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM task_check_items WHERE user_id = ?",
		"DELETE FROM task_comments WHERE user_id = ?",
		"DELETE FROM tasks WHERE user_id = ?",
		"DELETE FROM project_preferences WHERE user_id = ?",
		"DELETE FROM projects WHERE user_id = ?",
	} {
		if err := nullClearDB.Exec(stmt, uid).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// seedProjectRow 灌入清单行（archivedAt 非零 ⇒ 归档态），返回库中快照（供构造客户端 base）。
func seedProjectRow(t *testing.T, uid, pid int64, name string, archivedAt *time.Time) models.Project {
	t.Helper()
	base := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	p := &models.Project{
		ModelBase: models.ModelBase{ID: pid, CreatedAt: base, UpdatedAt: base},
		UserId:    uid,
		Name:      name,
	}
	if archivedAt != nil {
		p.ArchivedAt = sql.NullTime{Time: *archivedAt, Valid: true}
	}
	if err := nullClearDB.Create(p).Error; err != nil {
		t.Fatalf("seed 清单 %d 失败: %v", pid, err)
	}
	return readProjectRow(t, pid)
}

func readProjectRow(t *testing.T, pid int64) models.Project {
	t.Helper()
	var m models.Project
	if err := nullClearDB.Unscoped().First(&m, "id = ?", pid).Error; err != nil {
		t.Fatalf("读取清单 %d 失败: %v", pid, err)
	}
	return m
}

// projectPushBody 客户端同形清单推送载荷；archived 为附加字段 JSON 片段（可空）。
func projectPushBody(pid, uid int64, name, base string, archived string) string {
	body := fmt.Sprintf(`{"projects":[{"id":"%d","name":"%s","description":"","baseUpdatedAt":"%s"`,
		pid, name, base)
	if archived != "" {
		body += "," + archived
	}
	return body + `}]}`
}

// TestSyncProjectArchivedAtNullClears ① DEF-42 主场景：清单取消归档（`archivedAt: null`）⇒
// `archived_at` 变 NULL，且回执 applied（行级落地真的发生）。
func TestSyncProjectArchivedAtNullClears(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9801), int64(9802)
	cleanProjectNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRow(t, uid, pid, "T322 清单", &archivedAt)
	if !seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 清单应处于归档态")
	}

	body := projectPushBody(pid, uid, "T322 清单", idutil.FormatTimeMilli(seed.UpdatedAt), `"archivedAt":null`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readProjectRow(t, pid)
	if after.ArchivedAt.Valid {
		t.Fatalf("修复未生效：archived_at 仍为 %v（应被 null 清空）",
			after.ArchivedAt.Time.UTC().Format(time.RFC3339))
	}
	if !after.UpdatedAt.After(seed.UpdatedAt) {
		t.Fatalf("覆盖分支应 bump updated_at: %v → %v", seed.UpdatedAt, after.UpdatedAt)
	}
	t.Logf("① archivedAt:null ⇒ outcome=%s · archived_at %v → NULL · updated_at %v → %v",
		results[0].Outcome, archivedAt.UTC().Format(time.RFC3339),
		seed.UpdatedAt.UTC().Format(idutil.RFC3339Milli), after.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
}

// TestSyncProjectArchivedAtAbsentKeeps ② 真三态回归锁：缺省（旧客户端）⇒ 不清空，其余字段照常覆盖。
func TestSyncProjectArchivedAtAbsentKeeps(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9811), int64(9812)
	cleanProjectNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRow(t, uid, pid, "T322 清单", &archivedAt)

	body := projectPushBody(pid, uid, "T322 改名", idutil.FormatTimeMilli(seed.UpdatedAt), "")
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readProjectRow(t, pid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(seed.ArchivedAt.Time) {
		t.Fatalf("absent 不得改 archived_at: %v → %+v", seed.ArchivedAt.Time, after.ArchivedAt)
	}
	if after.Name != "T322 改名" {
		t.Fatalf("其它字段应照常覆盖: name = %q", after.Name)
	}
	t.Logf("② absent ⇒ archived_at 保持 %v · name → %s",
		after.ArchivedAt.Time.UTC().Format(time.RFC3339), after.Name)
}

// TestSyncProjectArchivedAtValueWrites ③ 归档：推送具体时间 ⇒ 写入该时间。
func TestSyncProjectArchivedAtValueWrites(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9821), int64(9822)
	cleanProjectNullClear(t, uid)
	seed := seedProjectRow(t, uid, pid, "T322 清单", nil)
	if seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 清单应为未归档态")
	}

	wantAt := time.Now().UTC().Truncate(time.Second).Add(-5 * time.Minute)
	body := projectPushBody(pid, uid, "T322 清单", idutil.FormatTimeMilli(seed.UpdatedAt),
		`"archivedAt":"`+wantAt.Format(time.RFC3339)+`"`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readProjectRow(t, pid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(wantAt) {
		t.Fatalf("archived_at = %+v, want %v", after.ArchivedAt, wantAt)
	}
	t.Logf("③ archivedAt=%s ⇒ archived_at 写入 %v", wantAt.Format(time.RFC3339),
		after.ArchivedAt.Time.UTC().Format(time.RFC3339))
}

// TestSyncProjectArchivedAtPullOutputUnchanged ④ pull 出参回归：清空后清单未归档表示为空串，
// 写入后为 RFC3339 秒级时间串。
func TestSyncProjectArchivedAtPullOutputUnchanged(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9831), int64(9832)
	cleanProjectNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRow(t, uid, pid, "T322 清单", &archivedAt)

	body := projectPushBody(pid, uid, "T322 清单", idutil.FormatTimeMilli(seed.UpdatedAt), `"archivedAt":null`)
	if _, raw := doNullClearPush(t, ctrl, uid, body); !strings.Contains(raw, `"outcome":"applied"`) {
		t.Fatalf("前置 push 未 applied: %s", raw)
	}

	ctx := context.Background()
	items, err := ctrl.projectApp.ListSync(ctx, uid, "", "", 10)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("pull 条数 = %d, want 1", len(items))
	}
	if items[0].ArchivedAt != "" {
		t.Fatalf("清空后 pull 出参 archivedAt = %q, want \"\"", items[0].ArchivedAt)
	}

	// 再写回归档值 ⇒ pull 出参为秒级时间串
	cur := readProjectRow(t, pid)
	body = projectPushBody(pid, uid, "T322 清单", idutil.FormatTimeMilli(cur.UpdatedAt),
		`"archivedAt":"`+archivedAt.Format(time.RFC3339)+`"`)
	if _, raw := doNullClearPush(t, ctrl, uid, body); !strings.Contains(raw, `"outcome":"applied"`) {
		t.Fatalf("回写归档未 applied: %s", raw)
	}
	items, err = ctrl.projectApp.ListSync(ctx, uid, "", "", 10)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if items[0].ArchivedAt != archivedAt.Format(time.RFC3339) {
		t.Fatalf("已归档 pull 出参 archivedAt = %q, want %q", items[0].ArchivedAt, archivedAt.Format(time.RFC3339))
	}
	t.Logf("④ pull 出参：未归档=%q · 已归档=%q", "", items[0].ArchivedAt)
}

// TestSyncProjectArchivedAtRestUnchanged ⑤ REST 契约回归：REST create / PATCH 不含 archived 语义 ⇒
// 已归档清单既不被清空也不被改写。
func TestSyncProjectArchivedAtRestUnchanged(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid = int64(9841), int64(9842)
	cleanProjectNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRow(t, uid, pid, "T322 清单", &archivedAt)
	if !seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 清单应处于归档态")
	}

	// REST create（共享 DTO，载荷里显式塞 archivedAt 也应被丢弃）
	rest := fmt.Sprintf(`{"name":"T322 REST","description":"","id":"%d","archivedAt":null}`, pid)
	var restReq types.CreateProjectReq
	if err := json.Unmarshal([]byte(rest), &restReq); err != nil {
		t.Fatalf("绑定 REST create 载荷: %v", err)
	}
	if _, _, err := ctrl.projectApp.Create(context.Background(), uid, toCreateProjectInput(&restReq)); err != nil {
		t.Fatalf("REST Create: %v", err)
	}
	after := readProjectRow(t, pid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(archivedAt) {
		t.Fatalf("REST create 不得改 archived_at: %v → %+v", archivedAt, after.ArchivedAt)
	}

	// REST PATCH（Name 更新）
	patchName := "T322 PATCH"
	if err := ctrl.projectApp.Update(context.Background(), uid, fmt.Sprintf("%d", pid), &projectDto.UpdateProjectReq{
		Name: &patchName,
	}); err != nil {
		t.Fatalf("REST Update: %v", err)
	}
	after = readProjectRow(t, pid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(archivedAt) {
		t.Fatalf("REST PATCH 不得改 archived_at: %v → %+v", archivedAt, after.ArchivedAt)
	}
	if after.Name != patchName {
		t.Fatalf("REST PATCH 未生效: name = %q", after.Name)
	}
	t.Logf("⑤ REST create/PATCH 后 archived_at 保持 %v · name → %s",
		after.ArchivedAt.Time.UTC().Format(time.RFC3339), after.Name)
}

// TestSyncProjectArchivedAtNoCascade ⑥ 协议约定锁：服务端**不级联** —— 只落地清单行；
// 清单下任务仍保持自身 archived_at（由客户端逐任务推送变更）。
func TestSyncProjectArchivedAtNoCascade(t *testing.T) {
	ctrl := newProjectSyncController()
	const uid, pid, tid = int64(9851), int64(9852), int64(9853)
	cleanProjectNullClear(t, uid)
	archivedAt := time.Now().UTC().Truncate(time.Second).Add(-30 * time.Minute)
	seed := seedProjectRow(t, uid, pid, "T322 清单", &archivedAt)

	// 同一清单下挂一条已归档任务（模拟仅清单被取消归档、任务尚未逐条推送）
	taskRepo := taskPersistence.NewTaskRepo(nullClearDB)
	base := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	vo := &taskValueObjects.CreateTask{
		Id: tid, Name: "T322 任务", CreatedAt: base, UpdatedAt: base,
		ProjectId:  domaintypes.ProjectID(pid),
		ArchivedAt: domaintypes.NewNullableTimeByTime(archivedAt),
	}
	if _, _, err := taskRepo.Upsert(context.Background(), uid, vo); err != nil {
		t.Fatalf("seed 任务失败: %v", err)
	}

	body := projectPushBody(pid, uid, "T322 清单", idutil.FormatTimeMilli(seed.UpdatedAt), `"archivedAt":null`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want applied；body = %s", results, raw)
	}
	projectAfter := readProjectRow(t, pid)
	if projectAfter.ArchivedAt.Valid {
		t.Fatalf("清单应被清空归档")
	}
	var taskAfter models.Task
	if err := nullClearDB.Unscoped().First(&taskAfter, "id = ?", tid).Error; err != nil {
		t.Fatalf("读取任务失败: %v", err)
	}
	if !taskAfter.ArchivedAt.Valid {
		t.Fatalf("服务端不得级联取消归档任务（协议约定：行级落地 + 客户端逐任务推送）")
	}
	t.Logf("⑥ 清单 archived_at=NULL，而任务 archived_at 保持 %v（未级联）",
		taskAfter.ArchivedAt.Time.UTC().Format(time.RFC3339))
}
