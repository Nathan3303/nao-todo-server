//go:build integration

// T327 集成测试（真实 MySQL + 完整装配：app + TxManager + 内存总线 + CountUpdater）：
// 服务端「派生写」回执（additive `derivedUpdates`），修复 RC-1
// （ADR 2026-09-28 sync-conflict-timestamp-basis：派生写推进父行 updated_at 却不回传
//
//	⇒ 客户端 base 静默过期 ⇒ 下次推父实体判 stale）。
//
// 红（未修）：回执不含 `derivedUpdates` ⇒ 父行被派生写推进后客户端 base 无法收敛。
// 绿（修后）：回执含被派生写行的最终版本，且与库中 updated_at 逐字相等。
//
// 复用同包 TestMain / nullClearDB / cleanNullClear / seedNullClearTask
// （见 sync_null_clear_integration_test.go）。运行：
//
//	TZ=UTC go test -tags integration -p 1 -count=1 -run TestSyncPushDerived -v ./interfaces/controllers/
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	counts "naotodoserver/application/counts"
	"naotodoserver/application/idutil"
	projectApp "naotodoserver/application/project"
	taskApp "naotodoserver/application/task"
	"naotodoserver/conf"
	projectService "naotodoserver/domain/project/service"
	taskService "naotodoserver/domain/task/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/events"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	projectPersistence "naotodoserver/infrastructure/persistence/project"
	taskPersistence "naotodoserver/infrastructure/persistence/task"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// derivedEcho 回执中的派生行更新段（测试本地结构，独立于生产类型：
// 便于「未修 = 字段缺省」直接判红，不因缺类型导致编译失败而丢失红证据）。
type derivedEcho struct {
	Table     string `json:"table"`
	Id        string `json:"id"`
	UpdatedAt string `json:"updatedAt"`
}

// derivedPushRes /sync/push 回执（含 additive 派生行段）。
type derivedPushRes struct {
	Results        []types.SyncResult `json:"results"`
	ServerTime     string             `json:"serverTime"`
	DerivedUpdates []derivedEcho      `json:"derivedUpdates"`
}

// newDerivedController 完整装配：真实仓储 + TxManager + 内存总线 + CountUpdater
// （与 production initialize.go 同构；Redis 不可达 ⇒ 缓存静默降级）。
func newDerivedController() *SyncController {
	repo := taskPersistence.NewTaskRepo(nullClearDB)
	projRepo := projectPersistence.NewProjectRepo(
		nullClearDB, cache.NewCache(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})),
	)
	prefRepo := projectPersistence.NewProjectPreferenceRepo(nullClearDB)
	bus := events.NewInMemoryBus()
	bus.Subscribe(counts.NewCountUpdater(repo, projRepo).HandleCountEvent)
	txManager := dbs.NewTxManager(nullClearDB)
	tApp := taskApp.NewTaskApp(
		taskService.NewTaskDomain(repo, repo), repo, repo, repo, nil, txManager, bus,
	)
	pApp := projectApp.NewProjectApp(
		projectService.NewProjectDomain(projRepo, prefRepo), txManager, projRepo, prefRepo, repo, bus,
	)
	return &SyncController{taskApp: tApp, checkItemApp: tApp, commentApp: tApp, projectApp: pApp}
}

// doDerivedPush 走完整 HTTP 绑定 + 控制器，返回含派生行段的回执。
func doDerivedPush(t *testing.T, ctrl *SyncController, uid int64, body string) derivedPushRes {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/sync/push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(iCtx.SetUserId(req.Context(), uid))
	gc.Request = req
	ctrl.Push(gc)
	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 状态 = %d, body = %s", w.Code, w.Body.String())
	}
	raw := w.Body.String()
	var envelope struct {
		Code int            `json:"code"`
		Data derivedPushRes `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatalf("解析响应失败: %v, body = %s", err, raw)
	}
	if envelope.Code != 90010 {
		t.Fatalf("回执 code = %d, body = %s", envelope.Code, raw)
	}
	return envelope.Data
}

// echoUpdatedAt 取回执中某 (table,id) 的派生行版本；缺省返回空串。
func echoUpdatedAt(resp derivedPushRes, table, id string) string {
	for _, d := range resp.DerivedUpdates {
		if d.Table == table && d.Id == id {
			return d.UpdatedAt
		}
	}
	return ""
}

// countEcho 统计回执中某 (table,id) 出现次数（去重断言用）。
func countEcho(resp derivedPushRes, table, id string) int {
	n := 0
	for _, d := range resp.DerivedUpdates {
		if d.Table == table && d.Id == id {
			n++
		}
	}
	return n
}

func strID(id int64) string { return strconv.FormatInt(id, 10) }

// taskPushBodyWithParent 客户端同形任务推送载荷（含 parentTaskId，空串表示顶层任务）。
func taskPushBodyWithParent(id, projectID, parentID int64, name, base string) string {
	body := fmt.Sprintf(`{"tasks":[{"id":"%d","name":"%s","description":"","state":"pending",`+
		`"priority":"medium","projectId":"%d","tags":[],"remindRepeat":"none","remindWeekdays":[]`,
		id, name, projectID)
	if parentID > 0 {
		body += fmt.Sprintf(`,"parentTaskId":"%d"`, parentID)
	}
	if base != "" {
		body += fmt.Sprintf(`,"baseUpdatedAt":"%s"`, base)
	}
	return body + `}]}`
}

// checkItemPushBody 客户端同形检查项推送载荷（base 为空表示由服务端自行判定/新建）。
func checkItemPushBody(id, taskID int64, name string, sortID uint16, base string) string {
	body := fmt.Sprintf(`{"taskCheckItems":[{"id":"%d","taskId":"%d","name":"%s",`+
		`"description":"","isDone":false,"sortId":%d`, id, taskID, name, sortID)
	if base != "" {
		body += fmt.Sprintf(`,"baseUpdatedAt":"%s"`, base)
	}
	return body + `}]}`
}

// readTaskRow 读任务行快照（复用同包 nullClearDB）。
func readTaskRow(t *testing.T, id int64) models.Task {
	t.Helper()
	return readNullClearTask(t, id)
}

// assertTaskEchoStrict 父/子任务行若被派生写推进 ⇒ 回执必须带回其最终库中版本（AC-T325-4 核心断言）。
func assertTaskEchoStrict(t *testing.T, step string, resp derivedPushRes, taskID int64, before time.Time) models.Task {
	t.Helper()
	after := readTaskRow(t, taskID)
	if !after.UpdatedAt.After(before) {
		t.Fatalf("[%s] 前置失败：任务 %d updated_at 未推进（before=%v, after=%v）",
			step, taskID, before.UTC().Format(idutil.RFC3339Milli), after.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	got := echoUpdatedAt(resp, "tasks", strID(taskID))
	if got == "" {
		t.Fatalf("[%s] 任务 %d updated_at 已推进 %v→%v，但回执 derivedUpdates 未带回该行版本 ⇒ 客户端 base 无法收敛（RC-1）",
			step, taskID,
			before.UTC().Format(idutil.RFC3339Milli), after.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	want := idutil.FormatTimeMilli(after.UpdatedAt.UTC())
	if got != want {
		t.Fatalf("[%s] 回执任务版本 %q != 库中 updated_at %q", step, got, want)
	}
	t.Logf("[%s] 任务 %d updated_at %v → %v，回执带回 %s ✓",
		step, taskID, before.UTC().Format(idutil.RFC3339Milli), want, got)
	return after
}

// assertProjectEchoStrict 项目行若被派生写推进 ⇒ 回执必须带回其最终库中版本。
func assertProjectEchoStrict(t *testing.T, step string, resp derivedPushRes, projectID int64, before time.Time) {
	t.Helper()
	after := readProjectRow(t, projectID)
	if !after.UpdatedAt.After(before) {
		t.Fatalf("[%s] 前置失败：项目 %d updated_at 未推进（before=%v, after=%v）",
			step, projectID, before.UTC().Format(idutil.RFC3339Milli), after.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	got := echoUpdatedAt(resp, "projects", strID(projectID))
	if got == "" {
		t.Fatalf("[%s] 项目 %d updated_at 已推进 %v→%v，但回执 derivedUpdates 未带回该行版本",
			step, projectID,
			before.UTC().Format(idutil.RFC3339Milli), after.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	want := idutil.FormatTimeMilli(after.UpdatedAt.UTC())
	if got != want {
		t.Fatalf("[%s] 回执项目版本 %q != 库中 updated_at %q", step, got, want)
	}
	t.Logf("[%s] 项目 %d updated_at %v → %v，回执带回 %s ✓",
		step, projectID, before.UTC().Format(idutil.RFC3339Milli), want, got)
}

// TestSyncPushDerivedUpdates_UserSequence 用户触发序列（逐字）：
//
//	改检查项名称 → 移动检查项 → 建子任务 → 改子任务属性
//
// 每一步断言父任务行 updated_at 是否被推进，以及（被推进时）回执是否带回父行版本。
// 红线：只有「计数联动」步骤（建子任务）会推进父行；rename/move/属性修改不触计数 ⇒ 不推进。
func TestSyncPushDerivedUpdates_UserSequence(t *testing.T) {
	ctrl := newDerivedController()
	const uid, pid, tid = int64(9901), int64(9902), int64(9903)
	const cid, sid = int64(9904), int64(9905)
	cleanNullClear(t, uid)

	repo := taskPersistence.NewTaskRepo(nullClearDB)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{})
	// 检查项直插（避免 pre-seed 计数事件污染基线）
	ci := models.TaskCheckItem{
		ModelBase: models.ModelBase{ID: cid, CreatedAt: seed.UpdatedAt, UpdatedAt: seed.UpdatedAt},
		UserId:    uid, TaskId: tid, Name: "检查项", SortId: 100,
	}
	if err := nullClearDB.Create(&ci).Error; err != nil {
		t.Fatalf("seed 检查项失败: %v", err)
	}
	baseline := readTaskRow(t, tid).UpdatedAt

	// 步骤 ①：改检查项名称（push upsert，created=false ⇒ 无计数联动）
	resp1 := doDerivedPush(t, ctrl, uid,
		checkItemPushBody(cid, tid, "检查项-改名", 100, idutil.FormatTimeMilli(ci.UpdatedAt)))
	if len(resp1.Results) != 1 || resp1.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("[改检查项名称] 回执 = %+v, want applied", resp1.Results)
	}
	after1 := readTaskRow(t, tid)
	if !after1.UpdatedAt.Equal(baseline) {
		t.Fatalf("[改检查项名称] 不应推进父任务 updated_at：%v → %v",
			baseline.UTC().Format(idutil.RFC3339Milli), after1.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	t.Logf("[改检查项名称] 父任务 updated_at 未推进（%v）· 回执派生段 = %v",
		after1.UpdatedAt.UTC().Format(idutil.RFC3339Milli), resp1.DerivedUpdates)

	// 步骤 ②：移动检查项（改 sortId，push upsert，created=false ⇒ 无计数联动）
	resp2 := doDerivedPush(t, ctrl, uid,
		checkItemPushBody(cid, tid, "检查项-改名", 500, resp1.Results[0].ServerUpdatedAt))
	if len(resp2.Results) != 1 || resp2.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("[移动检查项] 回执 = %+v, want applied", resp2.Results)
	}
	after2 := readTaskRow(t, tid)
	if !after2.UpdatedAt.Equal(baseline) {
		t.Fatalf("[移动检查项] 不应推进父任务 updated_at：%v → %v",
			baseline.UTC().Format(idutil.RFC3339Milli), after2.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	t.Logf("[移动检查项] 父任务 updated_at 未推进（%v）· 回执派生段 = %v",
		after2.UpdatedAt.UTC().Format(idutil.RFC3339Milli), resp2.DerivedUpdates)

	// 步骤 ③：创建子任务（subtask_count +1 ⇒ 父任务被派生写推进；项目 task_count +1）
	beforeSubtask := pid0UpdatedAt(t, pid)
	resp3 := doDerivedPush(t, ctrl, uid, taskPushBodyWithParent(sid, pid, tid, "子任务", ""))
	if len(resp3.Results) != 1 || resp3.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("[建子任务] 回执 = %+v, want applied", resp3.Results)
	}
	parentAfter3 := assertTaskEchoStrict(t, "建子任务", resp3, tid, baseline)
	assertProjectEchoStrict(t, "建子任务", resp3, pid, beforeSubtask)

	// 步骤 ④：修改子任务属性（改名，父未变 ⇒ 无计数联动）
	subtaskBase := idutil.FormatTimeMilli(readTaskRow(t, sid).UpdatedAt)
	resp4 := doDerivedPush(t, ctrl, uid,
		taskPushBodyWithParent(sid, pid, tid, "子任务-改名", subtaskBase))
	if len(resp4.Results) != 1 || resp4.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("[改子任务属性] 回执 = %+v, want applied", resp4.Results)
	}
	after4 := readTaskRow(t, tid)
	if !after4.UpdatedAt.Equal(parentAfter3.UpdatedAt) {
		t.Fatalf("[改子任务属性] 不应推进父任务 updated_at：%v → %v",
			parentAfter3.UpdatedAt.UTC().Format(idutil.RFC3339Milli), after4.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
	}
	t.Logf("[改子任务属性] 父任务 updated_at 未推进（%v）· 回执派生段 = %v",
		after4.UpdatedAt.UTC().Format(idutil.RFC3339Milli), resp4.DerivedUpdates)
}

// pid0UpdatedAt 读项目行当前 updated_at（步骤 ③ 前置基线）。
func pid0UpdatedAt(t *testing.T, projectID int64) time.Time {
	t.Helper()
	return readProjectRow(t, projectID).UpdatedAt
}

// TestSyncPushDerivedUpdates_SameParentDedupFinalVersion 同一父行在同批被多次派生写
// （建子任务 + 建两个检查项，同一父任务）⇒ 回执去重后**只给最终版本**，且与库中逐字相等。
func TestSyncPushDerivedUpdates_SameParentDedupFinalVersion(t *testing.T) {
	ctrl := newDerivedController()
	const uid, pid, tid = int64(9911), int64(9912), int64(9913)
	const sid, cidA, cidB = int64(9914), int64(9915), int64(9916)
	cleanNullClear(t, uid)

	repo := taskPersistence.NewTaskRepo(nullClearDB)
	seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{})
	baseline := readTaskRow(t, tid).UpdatedAt

	body := fmt.Sprintf(`{"tasks":[{"id":"%d","name":"子任务","description":"","state":"pending",`+
		`"priority":"medium","projectId":"%d","parentTaskId":"%d","tags":[],`+
		`"remindRepeat":"none","remindWeekdays":[]}],`+
		`"taskCheckItems":[`+
		`{"id":"%d","taskId":"%d","name":"检查项A","description":"","isDone":false,"sortId":100},`+
		`{"id":"%d","taskId":"%d","name":"检查项B","description":"","isDone":false,"sortId":200}]}`,
		sid, pid, tid, cidA, tid, cidB, tid)
	resp := doDerivedPush(t, ctrl, uid, body)
	if len(resp.Results) != 3 {
		t.Fatalf("回执条数 = %d, want 3；%+v", len(resp.Results), resp.Results)
	}

	after := readTaskRow(t, tid)
	if !after.UpdatedAt.After(baseline) {
		t.Fatalf("同批建子任务+检查项应推进父任务 updated_at：%v → %v", baseline, after.UpdatedAt)
	}
	if n := countEcho(resp, "tasks", strID(tid)); n != 1 {
		t.Fatalf("同一父行同批多次派生写 ⇒ 回执应去重为 1 条，实际 %d 条：%+v", n, resp.DerivedUpdates)
	}
	want := idutil.FormatTimeMilli(after.UpdatedAt.UTC())
	if got := echoUpdatedAt(resp, "tasks", strID(tid)); got != want {
		t.Fatalf("去重后应给最终版本：回执 %q != 库中 %q", got, want)
	}
	// 计数联动本身也要落地（去重不掩盖多次写）
	if after.SubtaskCount != 1 || after.CheckItemCount != 2 {
		t.Fatalf("计数联动异常：subtask_count=%d check_item_count=%d, want 1/2",
			after.SubtaskCount, after.CheckItemCount)
	}
	t.Logf("同批 3 次派生写 ⇒ 回执去重 1 条 = 库中 %s · subtask_count=%d check_item_count=%d",
		want, after.SubtaskCount, after.CheckItemCount)
}

// TestSyncPushDerivedUpdates_SingleEndRA AC-T325-5（R-A 端到端，服务端侧等价探针）：
// 改名 T → 加检查项 → 再改名 T。对照组证明「不收敛 ⇒ stale」（RC-1 红）；
// 修复组用回执带回的父行版本收敛 base ⇒ applied 且服务端=最后一次改名（冲突 +0）。
func TestSyncPushDerivedUpdates_SingleEndRA(t *testing.T) {
	ctrl := newDerivedController()
	const uid, pid, tid = int64(9921), int64(9922), int64(9923)
	const cid = int64(9924)
	cleanNullClear(t, uid)

	repo := taskPersistence.NewTaskRepo(nullClearDB)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{})
	base0 := idutil.FormatTimeMilli(seed.UpdatedAt)

	// ① 改名 T（base = 库中版本）⇒ applied
	resp1 := doDerivedPush(t, ctrl, uid, taskPushBodyWithParent(tid, pid, 0, "T-v1", base0))
	if len(resp1.Results) != 1 || resp1.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("① 改名 T：回执 = %+v, want applied", resp1.Results)
	}
	base1 := resp1.Results[0].ServerUpdatedAt

	// ② 加检查项 ⇒ 服务端派生写推进父任务；回执必须带回 S2
	resp2 := doDerivedPush(t, ctrl, uid, checkItemPushBody(cid, tid, "检查项", 100, ""))
	if len(resp2.Results) != 1 || resp2.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("② 加检查项：回执 = %+v, want applied", resp2.Results)
	}
	parentS2 := readTaskRow(t, tid).UpdatedAt
	base2Echo := echoUpdatedAt(resp2, "tasks", strID(tid))
	if base2Echo == "" {
		t.Fatalf("② 父任务被派生写推进，但回执未带回其版本 ⇒ 客户端 base 无法收敛（RC-1 红）")
	}
	if want := idutil.FormatTimeMilli(parentS2.UTC()); base2Echo != want {
		t.Fatalf("② 回执父任务版本 %q != 库中 %q", base2Echo, want)
	}

	// ③ 对照组（红证据）：仍用陈旧 base1 改名 ⇒ 服务端判 stale（不写入）
	resp3a := doDerivedPush(t, ctrl, uid, taskPushBodyWithParent(tid, pid, 0, "T-stale-should-not-write", base1))
	if len(resp3a.Results) != 1 || resp3a.Results[0].Outcome != types.SyncOutcomeStale {
		t.Fatalf("③ 对照组应 stale（陈旧 base），回执 = %+v", resp3a.Results)
	}
	if got := readTaskRow(t, tid).Name; got != "T-v1" {
		t.Fatalf("③ 对照组 stale 不得写入：name = %q, want T-v1", got)
	}

	// ④ 修复组（绿证据）：用回执带回的 base2 收敛后改名 ⇒ applied，服务端=最后一次改名
	resp3b := doDerivedPush(t, ctrl, uid, taskPushBodyWithParent(tid, pid, 0, "T-v2", base2Echo))
	if len(resp3b.Results) != 1 || resp3b.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("④ 收敛 base 后改名应 applied（冲突 +0），回执 = %+v", resp3b.Results)
	}
	if got := readTaskRow(t, tid).Name; got != "T-v2" {
		t.Fatalf("④ 服务端最终名称 = %q, want T-v2", got)
	}
	t.Logf("R-A：base0=%s → 改名 applied → 检查项派生 S2=%s（回执带回）→ 陈旧 base 判 stale（红）→ 收敛 base 改名 applied，服务端=T-v2（绿）",
		base0, base2Echo)
}

// ensureDerivedUser 评论创建需用户昵称/头像快照（缺失时补一行）。
func ensureDerivedUser(t *testing.T, uid int64) {
	t.Helper()
	if conf.Conf == nil {
		conf.Conf = &conf.Config{Uploads: &conf.Uploads{}}
	}
	var cnt int64
	if err := nullClearDB.Model(&models.User{}).Where("id = ?", uid).Count(&cnt).Error; err != nil {
		t.Fatalf("查询用户: %v", err)
	}
	if cnt > 0 {
		return
	}
	if err := nullClearDB.Create(&models.User{
		ModelBase: models.ModelBase{ID: uid},
		Account:   fmt.Sprintf("t327_%d", uid),
		Email:     fmt.Sprintf("t327_%d@example.com", uid),
		Password:  "x",
		Nickname:  "t327",
	}).Error; err != nil {
		t.Fatalf("插入用户: %v", err)
	}
}

// TestSyncPushDerivedUpdates_CommentAndCheckItemDelete 其余子对象写点的父行联动：
// 评论创建（comment_count +1）/ 检查项删除（check_item_count -1）⇒ 父任务行被派生写推进并回传。
func TestSyncPushDerivedUpdates_CommentAndCheckItemDelete(t *testing.T) {
	ctrl := newDerivedController()
	const uid, pid, tid = int64(9941), int64(9942), int64(9943)
	const cid, cid2 = int64(9944), int64(9945)
	cleanNullClear(t, uid)
	ensureDerivedUser(t, uid)

	repo := taskPersistence.NewTaskRepo(nullClearDB)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{})
	ci := models.TaskCheckItem{
		ModelBase: models.ModelBase{ID: cid, CreatedAt: seed.UpdatedAt, UpdatedAt: seed.UpdatedAt},
		UserId:    uid, TaskId: tid, Name: "待删检查项", SortId: 100,
	}
	if err := nullClearDB.Create(&ci).Error; err != nil {
		t.Fatalf("seed 检查项: %v", err)
	}
	// 直插检查项不同步计数列 ⇒ 手动对齐（UpdateColumn 不自动 bump updated_at）
	if err := nullClearDB.Model(&models.Task{}).Where("id = ?", tid).
		UpdateColumn("check_item_count", 1).Error; err != nil {
		t.Fatalf("对齐 check_item_count: %v", err)
	}
	base := readTaskRow(t, tid).UpdatedAt

	// 评论创建 ⇒ 父任务 comment_count 联动
	resp1 := doDerivedPush(t, ctrl, uid,
		fmt.Sprintf(`{"taskComments":[{"id":"%d","taskId":"%d","content":"c"}]}`, cid2, tid))
	if len(resp1.Results) != 1 || resp1.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("[评论创建] 回执 = %+v, want applied", resp1.Results)
	}
	after1 := assertTaskEchoStrict(t, "评论创建", resp1, tid, base)

	// 检查项删除 ⇒ 父任务 check_item_count 联动
	resp2 := doDerivedPush(t, ctrl, uid,
		fmt.Sprintf(`{"deletions":[{"table":"taskCheckItems","id":"%d"}]}`, cid))
	if len(resp2.Results) != 1 || resp2.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("[检查项删除] 回执 = %+v, want applied", resp2.Results)
	}
	assertTaskEchoStrict(t, "检查项删除", resp2, tid, after1.UpdatedAt)
}

// TestSyncPushDerivedUpdates_ImplicitBucketNoPhantom 隐式桶（projectId="" ⇒ userId 虚拟项目，
// 无 projects 行）：任务创建必须**不**回传幻影项目版本（否则客户端会以不存在的版本作 base）。
func TestSyncPushDerivedUpdates_ImplicitBucketNoPhantom(t *testing.T) {
	ctrl := newDerivedController()
	const uid = int64(9951)
	const sid = int64(9952)
	cleanNullClear(t, uid)

	resp := doDerivedPush(t, ctrl, uid,
		fmt.Sprintf(`{"tasks":[{"id":"%d","name":"收件箱任务","description":"",`+
			`"state":"pending","priority":"medium","projectId":"","tags":[],`+
			`"remindRepeat":"none","remindWeekdays":[]}]}`, sid))
	if len(resp.Results) != 1 || resp.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("隐式桶建任务：回执 = %+v, want applied", resp.Results)
	}
	if got := echoUpdatedAt(resp, "projects", strID(uid)); got != "" {
		t.Fatalf("隐式桶无 projects 行，不得回传幻影项目版本 %q：%+v", got, resp.DerivedUpdates)
	}
	t.Logf("隐式桶建任务：派生段 = %+v（无 projects 幻影）", resp.DerivedUpdates)
}

// TestSyncPushDerivedUpdates_ProjectDeleteCascade 项目删除（push deletions）⇒ 级联软删任务：
// 每个被触及任务行 + 重算后的项目行版本都必须回传（否则客户端对任一行 base 静默过期）。
func TestSyncPushDerivedUpdates_ProjectDeleteCascade(t *testing.T) {
	ctrl := newDerivedController()
	const uid, pid = int64(9961), int64(9962)
	const t1, t2 = int64(9963), int64(9964)
	cleanNullClear(t, uid)

	repo := taskPersistence.NewTaskRepo(nullClearDB)
	seedNullClearTask(t, repo, uid, t1, pid, nullClearSeed{})
	seedNullClearTask(t, repo, uid, t2, pid, nullClearSeed{})
	projectBase := readProjectRow(t, pid).UpdatedAt

	resp := doDerivedPush(t, ctrl, uid,
		fmt.Sprintf(`{"deletions":[{"table":"projects","id":"%d"}]}`, pid))
	if len(resp.Results) != 1 || resp.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("项目删除：回执 = %+v, want applied", resp.Results)
	}
	for _, tid := range []int64{t1, t2} {
		task := readTaskRow(t, tid)
		if !task.DeletedAt.Valid {
			t.Fatalf("级联：任务 %d 应已软删", tid)
		}
		want := idutil.FormatTimeMilli(task.UpdatedAt.UTC())
		if got := echoUpdatedAt(resp, "tasks", strID(tid)); got != want {
			t.Fatalf("级联任务 %d 回执 %q != 库中 %q", tid, got, want)
		}
	}
	if got := echoUpdatedAt(resp, "projects", strID(pid)); got == "" {
		t.Fatalf("项目删除重算应回传项目行版本：%+v", resp.DerivedUpdates)
	} else if want := idutil.FormatTimeMilli(readProjectRow(t, pid).UpdatedAt.UTC()); got != want {
		t.Fatalf("项目回执 %q != 库中 %q", got, want)
	}
	if !readProjectRow(t, pid).UpdatedAt.After(projectBase) {
		t.Fatalf("项目删除应推进项目 updated_at")
	}
	t.Logf("项目删除级联：tasks %d/%d + projects %d 均回传最终版本 ✓", t1, t2, pid)
}

// TestSyncPushDerivedUpdates_OCCDeploymentSelfCheck AC-T325-6（部署版本自检）：
// 本集成测试与 OCC 代码同仓编译 ⇒ OCC 必然在二进制内；断言「base 不匹配 ⇒ stale 且
// serverUpdatedAt = 库中当前版本」，据此证明运行端含 OCC（无需 skip，因为不可能在旧二进制上跑）。
func TestSyncPushDerivedUpdates_OCCDeploymentSelfCheck(t *testing.T) {
	ctrl := newDerivedController()
	const uid, pid, tid = int64(9931), int64(9932), int64(9933)
	cleanNullClear(t, uid)

	repo := taskPersistence.NewTaskRepo(nullClearDB)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{})

	staleBase := idutil.FormatTimeMilli(seed.UpdatedAt.Add(-time.Hour))
	resp := doDerivedPush(t, ctrl, uid, taskPushBodyWithParent(tid, pid, 0, "不应写入", staleBase))
	if len(resp.Results) != 1 || resp.Results[0].Outcome != types.SyncOutcomeStale {
		t.Fatalf("OCC 自检失败：base 不匹配应 stale（该构建含 OCC），回执 = %+v", resp.Results)
	}
	if got, want := resp.Results[0].ServerUpdatedAt, idutil.FormatTimeMilli(readTaskRow(t, tid).UpdatedAt.UTC()); got != want {
		t.Fatalf("stale 应回传库中当前版本：got %q want %q", got, want)
	}
	t.Logf("OCC 自检：stale 可达，serverUpdatedAt=%s = 库中版本 ✓", resp.Results[0].ServerUpdatedAt)
}
