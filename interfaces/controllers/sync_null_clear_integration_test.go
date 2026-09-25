//go:build integration

// T319 集成测试（真实 MySQL）：/sync/push 显式清空可空时间字段（修复 T318 P1 同步回滚）。
//
// 与 T318 探针同强度，但走**完整控制面链路**：HTTP JSON 载荷 → gin 绑定 → 控制器转换 →
// 真实应用层 → 真实仓储 → MySQL 列断言 + 返回体 outcome 断言。
//
//	启动：docker run -d --name t319-mysql -e MYSQL_ROOT_PASSWORD=dev_password \
//	  -e MYSQL_DATABASE=nao_todo_test -p 3307:3306 mysql:8.4.9
//	运行：TZ=UTC go test -tags integration -count=1 -run TestSyncNullClear -v ./interfaces/controllers/
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"naotodoserver/application/idutil"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	"naotodoserver/domain/task/entities"
	taskService "naotodoserver/domain/task/service"
	taskValueObjects "naotodoserver/domain/task/valueobjects"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	taskPersistence "naotodoserver/infrastructure/persistence/task"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

var nullClearDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("NAO_TEST_MYSQL_DSN")
	if dsn == "" {
		dsn = "root:dev_password@tcp(127.0.0.1:3307)/nao_todo_test?parseTime=true&loc=UTC&charset=utf8mb4"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Printf("连接测试数据库失败（需先启动测试 MySQL）: %v\n", err)
		os.Exit(1)
	}
	if err := db.AutoMigrate(
		&models.Project{}, &models.ProjectPreference{},
		&models.Task{}, &models.TaskCheckItem{}, &models.TaskComment{},
	); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}
	models.InitSnowflake(1)
	nullClearDB = db
	os.Exit(m.Run())
}

// newNullClearController 真实装配：仓储（MySQL）→ 领域 → 应用 → sync 控制器。
// 其余表应用传 nil：本用例载荷不含那些表，Push 不会触碰（列表为空 ⇒ 零次调用）。
func newNullClearController() (*SyncController, *taskPersistence.TaskRepoImpl) {
	repo := taskPersistence.NewTaskRepo(nullClearDB)
	domain := taskService.NewTaskDomain(repo, repo)
	app := taskApp.NewTaskApp(domain, repo, repo, repo, nil, dbs.NewTxManager(nullClearDB), nil)
	return &SyncController{taskApp: app}, repo
}

// cleanNullClear 清理本用例用户的数据（外键：偏好先删）。
func cleanNullClear(t *testing.T, uid int64) {
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

type nullClearSeed struct {
	archived bool
	starred  bool
	reminded bool
}

// seedNullClearTask 灌入一条任务并返回当前库中行快照（用于构造客户端 base）。
func seedNullClearTask(
	t *testing.T, repo *taskPersistence.TaskRepoImpl, uid, id, projectID int64, seed nullClearSeed,
) models.Task {
	t.Helper()
	ctx := context.Background()
	base := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	if err := nullClearDB.Create(&models.Project{
		ModelBase: models.ModelBase{ID: projectID, CreatedAt: base, UpdatedAt: base},
		UserId:    uid,
		Name:      "T319 清单",
	}).Error; err != nil && !strings.Contains(err.Error(), "Duplicate") {
		t.Fatalf("seed 清单失败: %v", err)
	}
	vo := &taskValueObjects.CreateTask{
		Id:        id,
		Name:      "T319 任务",
		CreatedAt: base,
		UpdatedAt: base,
		ProjectId: domaintypes.ProjectID(projectID),
		State:     entities.TaskStatePending,
		Priority:  entities.TaskPriorityMedium,
	}
	if seed.archived {
		vo.ArchivedAt = domaintypes.NewNullableTimeByTime(base.Add(10 * time.Minute))
	}
	if seed.starred {
		vo.StarMarkAt = domaintypes.NewNullableTimeByTime(base.Add(20 * time.Minute))
	}
	if seed.reminded {
		vo.RemindAt = domaintypes.NewNullableTimeByTime(base.Add(30 * time.Minute))
	}
	if _, _, err := repo.Upsert(ctx, uid, vo); err != nil {
		t.Fatalf("seed 任务 %d 失败: %v", id, err)
	}
	var m models.Task
	if err := nullClearDB.Unscoped().First(&m, "id = ?", id).Error; err != nil {
		t.Fatalf("读取任务 %d 失败: %v", id, err)
	}
	return m
}

// doNullClearPush 走完整 HTTP 绑定 + 控制器：返回回执 results 与原始响应体。
func doNullClearPush(
	t *testing.T, ctrl *SyncController, uid int64, body string,
) ([]types.SyncResult, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/sync/push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req.WithContext(iCtx.SetUserId(req.Context(), uid))
	ctrl.Push(c)

	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 状态 = %d, body = %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Code int               `json:"code"`
		Data types.SyncPushRes `json:"data"`
	}
	raw := w.Body.String()
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatalf("解析响应失败: %v, body = %s", err, raw)
	}
	if envelope.Code != 90010 {
		t.Fatalf("回执 code = %d, body = %s", envelope.Code, raw)
	}
	return envelope.Data.Results, raw
}

// taskPushBody 构造客户端同形任务推送载荷；nullable 为附加的可空时间字段 JSON 片段（可空）。
func taskPushBody(id int64, uid, projectID int64, name string, base string, nullable string) string {
	body := fmt.Sprintf(`{"tasks":[{"id":"%d","name":"%s","description":"","state":"pending",`+
		`"priority":"medium","projectId":"%d","tags":[],"remindRepeat":"none","remindWeekdays":[],`+
		`"baseUpdatedAt":"%s"`, id, name, projectID, base)
	if nullable != "" {
		body += "," + nullable
	}
	return body + `}]}`
}

func readNullClearTask(t *testing.T, id int64) models.Task {
	t.Helper()
	var m models.Task
	if err := nullClearDB.Unscoped().First(&m, "id = ?", id).Error; err != nil {
		t.Fatalf("读取任务 %d 失败: %v", id, err)
	}
	return m
}

// TestSyncNullClearArchivedAt 复现 T318 Q2 并证明修复：客户端发 `archivedAt: null`
// ⇒ 服务端 archived_at 变 NULL，且回执仍为 applied（写确实落地，不是被 stale/noop 挡下）。
func TestSyncNullClearArchivedAt(t *testing.T) {
	ctrl, repo := newNullClearController()
	const uid, pid, tid = int64(9701), int64(9702), int64(9703)
	cleanNullClear(t, uid)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{archived: true})
	if !seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 任务应处于归档态")
	}

	body := taskPushBody(tid, uid, pid, "T319 任务", idutil.FormatTimeMilli(seed.UpdatedAt), `"archivedAt":null`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readNullClearTask(t, tid)
	if after.ArchivedAt.Valid {
		t.Fatalf("修复未生效：archived_at 仍为 %v（应被 null 清空）",
			after.ArchivedAt.Time.UTC().Format(time.RFC3339))
	}
	if after.Name != "T319 任务" || !after.UpdatedAt.After(seed.UpdatedAt) {
		t.Fatalf("覆盖分支应同时写入其它列并 bump updated_at: name=%q updated_at=%v",
			after.Name, after.UpdatedAt)
	}
	t.Logf("修复后：outcome=%s · archived_at %v → NULL · updated_at %v → %v",
		results[0].Outcome, seed.ArchivedAt.Time.UTC().Format(time.RFC3339),
		seed.UpdatedAt.UTC().Format(idutil.RFC3339Milli), after.UpdatedAt.UTC().Format(idutil.RFC3339Milli))
}

// TestSyncNullClearAbsentKeepsArchived 真三态回归锁：载荷**不含** archivedAt 键（absent）
// ⇒ 不得清空（archived_at 保持原值），其余字段照常覆盖。
func TestSyncNullClearAbsentKeepsArchived(t *testing.T) {
	ctrl, repo := newNullClearController()
	const uid, pid, tid = int64(9711), int64(9712), int64(9713)
	cleanNullClear(t, uid)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{archived: true})

	body := taskPushBody(tid, uid, pid, "T319 改名", idutil.FormatTimeMilli(seed.UpdatedAt), "")
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readNullClearTask(t, tid)
	if !after.ArchivedAt.Valid || !after.ArchivedAt.Time.Equal(seed.ArchivedAt.Time) {
		t.Fatalf("absent 不得改 archived_at: %v → %v", seed.ArchivedAt.Time, after.ArchivedAt)
	}
	if after.Name != "T319 改名" {
		t.Fatalf("其它字段应照常覆盖: name = %q", after.Name)
	}
	t.Logf("absent：archived_at 保持 %v（未被清空）· name → %s",
		after.ArchivedAt.Time.UTC().Format(time.RFC3339), after.Name)
}

// TestSyncNullClearStarAndRemindAt 复现同 class 其余字段：starMarkAt / remindAt 发 null ⇒ 清空。
func TestSyncNullClearStarAndRemindAt(t *testing.T) {
	ctrl, repo := newNullClearController()
	const uid, pid, tid = int64(9721), int64(9722), int64(9723)
	cleanNullClear(t, uid)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{archived: true, starred: true, reminded: true})
	if !seed.StarMarkAt.Valid || !seed.RemindAt.Valid {
		t.Fatalf("前置失败：star/remind 应已置位")
	}

	body := taskPushBody(tid, uid, pid, "T319 任务", idutil.FormatTimeMilli(seed.UpdatedAt),
		`"starMarkAt":null,"remindAt":null`)
	results, raw := doNullClearPush(t, ctrl, uid, body)
	if len(results) != 1 || results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("回执 = %+v, want 单条 applied；body = %s", results, raw)
	}
	after := readNullClearTask(t, tid)
	if after.StarMarkAt.Valid || after.RemindAt.Valid {
		t.Fatalf("star/remind 未被清空: star=%v remind=%v", after.StarMarkAt, after.RemindAt)
	}
	if !after.ArchivedAt.Valid {
		t.Fatalf("未提供的 archivedAt（absent）被误清空")
	}
	t.Logf("star/remind null ⇒ 两列均 NULL；absent 的 archived_at 保持 %v",
		after.ArchivedAt.Time.UTC().Format(time.RFC3339))
}

// TestSyncNullClearPullOutputUnchanged pull 出参回归：清空后 pull 该行 archivedAt 仍输出 ""。
func TestSyncNullClearPullOutputUnchanged(t *testing.T) {
	ctrl, repo := newNullClearController()
	const uid, pid, tid = int64(9731), int64(9732), int64(9733)
	cleanNullClear(t, uid)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{archived: true})
	body := taskPushBody(tid, uid, pid, "T319 任务", idutil.FormatTimeMilli(seed.UpdatedAt), `"archivedAt":null`)
	if _, raw := doNullClearPush(t, ctrl, uid, body); !strings.Contains(raw, `"outcome":"applied"`) {
		t.Fatalf("前置 push 未 applied: %s", raw)
	}

	ctx := context.Background()
	items, err := ctrl.taskApp.ListTaskSync(ctx, uid, &taskDto.ListTaskReq{Limit: 10})
	if err != nil {
		t.Fatalf("ListTaskSync: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("pull 条数 = %d, want 1", len(items))
	}
	if items[0].ArchivedAt != "" {
		t.Fatalf("pull 出参 archivedAt = %q, want \"\"（客户端 '' → null）", items[0].ArchivedAt)
	}
	if items[0].UpdatedAt == "" {
		t.Fatalf("pull 出参 updatedAt 不应为空")
	}
	t.Logf("pull 回归：archivedAt=%q · updatedAt=%s", items[0].ArchivedAt, items[0].UpdatedAt)
}

// TestSyncNullClearRestCreateUnchanged REST create 契约回归：共享 create DTO 的
// `archivedAt: null` 仍是「缺省不写列」⇒ 已归档任务不被清空（本单未改 REST 语义）。
func TestSyncNullClearRestCreateUnchanged(t *testing.T) {
	ctrl, repo := newNullClearController()
	const uid, pid, tid = int64(9741), int64(9742), int64(9743)
	cleanNullClear(t, uid)
	seed := seedNullClearTask(t, repo, uid, tid, pid, nullClearSeed{archived: true})
	if !seed.ArchivedAt.Valid {
		t.Fatalf("前置失败：seed 任务应处于归档态")
	}

	// REST 载荷（create 契约，不含 baseUpdatedAt/updatedAt ⇒ LWW 走 Overwrite）
	rest := fmt.Sprintf(`{"id":"%d","name":"T319 REST","state":"pending","priority":"medium",`+
		`"projectId":"%d","archivedAt":null}`, tid, pid)
	var restReq types.CreateTaskReq
	if err := json.Unmarshal([]byte(rest), &restReq); err != nil {
		t.Fatalf("绑定 REST 载荷: %v", err)
	}
	if _, _, err := ctrl.taskApp.CreateTask(context.Background(), uid, toCreateTaskReq(&restReq)); err != nil {
		t.Fatalf("REST CreateTask: %v", err)
	}
	after := readNullClearTask(t, tid)
	if !after.ArchivedAt.Valid {
		t.Fatalf("REST create 的 null 不应清空 archived_at（契约回归）")
	}
	if after.Name != "T319 REST" {
		t.Fatalf("REST 覆盖未生效: name = %q", after.Name)
	}
	t.Logf("REST create 回归：archived_at 保持 %v · name → %s",
		after.ArchivedAt.Time.UTC().Format(time.RFC3339), after.Name)
}
