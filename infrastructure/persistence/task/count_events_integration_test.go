//go:build integration

// 领域统计属性联动 —— 全链路集成测试（真实 MySQL + 完整装配：app + 内存总线 + CountUpdater）
// 覆盖 ADR §12 server：U-S1 写路径×计数、U-S2 B2 updated_at 前进、U-S3 LWW/created/复活、
// U-S4 move/换父/Copy 双事件、同事务原子性（计数失败 ⇒ 主写回滚）。
// 依赖 repoImpl_integration_test.go 的 TestMain（已迁移 Task/Project/CheckItem/Comment/ProjectPreference）。
package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	counts "naotodoserver/application/counts"
	"naotodoserver/application/idutil"
	projectApp "naotodoserver/application/project"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	"naotodoserver/conf"
	projService "naotodoserver/domain/project/service"
	"naotodoserver/domain/task/entities"
	taskService "naotodoserver/domain/task/service"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/events"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	projRepo "naotodoserver/infrastructure/persistence/project"

	"github.com/go-redis/redis/v8"
)

const (
	testUserID int64 = 424242
	testProjA  int64 = 10001
	testProjB  int64 = 10002
)

// fullStack 完整装配：真实仓储 + TxManager + 内存总线 + CountUpdater
type fullStack struct {
	taskApp *taskApp.TaskAppImpl
	projApp projectApp.ProjectApp
	bus     *events.InMemoryBus
}

// newFullStack 装配全链路（Redis 不可用 ⇒ 项目列表缓存降级未命中，不影响计数断言）
func newFullStack(t *testing.T) *fullStack {
	t.Helper()
	// 评论响应转换需要 conf.Conf（AvatarURL）；测试注入最小配置避免依赖 config.yaml
	if conf.Conf == nil {
		conf.Conf = &conf.Config{Uploads: &conf.Uploads{}}
	}
	taskRepoInst := NewTaskRepo(testDB)
	projRepoInst := projRepo.NewProjectRepo(
		testDB,
		cache.NewCache(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})),
	)
	bus := events.NewInMemoryBus()
	updater := counts.NewCountUpdater(taskRepoInst, projRepoInst)
	bus.Subscribe(updater.HandleCountEvent)
	txManager := dbs.NewTxManager(testDB)
	taskDomain := taskService.NewTaskDomain(taskRepoInst, taskRepoInst)
	projDomain := projService.NewProjectDomain(projRepoInst, projRepo.NewProjectPreferenceRepo(testDB))
	tApp := taskApp.NewTaskApp(taskDomain, taskRepoInst, taskRepoInst, taskRepoInst, nil, txManager, bus)
	pApp := projectApp.NewProjectApp(projDomain, txManager, projRepoInst, projRepo.NewProjectPreferenceRepo(testDB), taskRepoInst, bus)
	return &fullStack{taskApp: tApp, projApp: pApp, bus: bus}
}

// insertProject 直插项目行（绕过偏好装配；计数调整只读该项目行）
func insertProject(t *testing.T, id int64) {
	t.Helper()
	if err := testDB.Create(&models.Project{
		ModelBase: models.ModelBase{ID: id},
		UserId:    testUserID,
		Name:      "p" + strconv.FormatInt(id, 10),
	}).Error; err != nil {
		t.Fatalf("插入项目 %d: %v", id, err)
	}
}

// ensureTestUser 直插测试用户行（评论创建需用户昵称/头像快照）
func ensureTestUser(t *testing.T) {
	t.Helper()
	var cnt int64
	if err := testDB.Model(&models.User{}).Where("id = ?", testUserID).Count(&cnt).Error; err != nil {
		t.Fatalf("查询测试用户: %v", err)
	}
	if cnt > 0 {
		return
	}
	if err := testDB.Create(&models.User{
		ModelBase: models.ModelBase{ID: testUserID},
		Account:   "count_test_user",
		Email:     "count_test@example.com",
		Password:  "hashed",
		Nickname:  "count-test",
	}).Error; err != nil {
		t.Fatalf("插入测试用户: %v", err)
	}
}

// taskReq 构造创建任务请求（项目/父任务 ID 均有效；state/priority 使用合法枚举串）
func taskReq(name string, projectID, parentTaskID int64, id *string) *taskDto.CreateTaskReq {
	req := &taskDto.CreateTaskReq{
		Name:         name,
		State:        "todo",
		Priority:     "medium",
		ProjectId:    idutil.FormatID(projectID),
		ParentTaskId: "",
		Id:           id,
	}
	if parentTaskID > 0 {
		req.ParentTaskId = idutil.FormatID(parentTaskID)
	}
	return req
}

// createTask 经 app 创建任务并返回解析后的 int64 ID
func createTask(t *testing.T, s *fullStack, req *taskDto.CreateTaskReq) int64 {
	t.Helper()
	res, err := s.taskApp.CreateTask(context.Background(), testUserID, req)
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	id, err := idutil.ParseID(res.Id)
	if err != nil {
		t.Fatalf("解析任务 ID %q: %v", res.Id, err)
	}
	return id
}

// taskCountsOf 读任务三计数列
func taskCountsOf(t *testing.T, taskID int64) (checkItems, comments, subtasks uint) {
	t.Helper()
	var m models.Task
	if err := testDB.Unscoped().Where("id = ?", taskID).First(&m).Error; err != nil {
		t.Fatalf("读任务 %d: %v", taskID, err)
	}
	return m.CheckItemCount, m.CommentCount, m.SubtaskCount
}

// projectTaskCount 读项目 task_count
func projectTaskCount(t *testing.T, projectID int64) uint {
	t.Helper()
	var m models.Project
	if err := testDB.Unscoped().Where("id = ?", projectID).First(&m).Error; err != nil {
		t.Fatalf("读项目 %d: %v", projectID, err)
	}
	return m.TaskCount
}

// taskUpdatedAt 读任务 updated_at
func taskUpdatedAt(t *testing.T, taskID int64) time.Time {
	t.Helper()
	var m models.Task
	if err := testDB.Unscoped().Where("id = ?", taskID).First(&m).Error; err != nil {
		t.Fatalf("读任务 %d: %v", taskID, err)
	}
	return m.UpdatedAt
}

// projectUpdatedAt 读项目 updated_at
func projectUpdatedAt(t *testing.T, projectID int64) time.Time {
	t.Helper()
	var m models.Project
	if err := testDB.Unscoped().Where("id = ?", projectID).First(&m).Error; err != nil {
		t.Fatalf("读项目 %d: %v", projectID, err)
	}
	return m.UpdatedAt
}

// backdateTask 回拨任务 updated_at 5 秒并返回真实落库值（秒级列 ⇒ 5s 间隔保证后续 bump 严格前进可判）
// 返回值即「客户端此前已拉到的版本」，可作 keyset 游标。
func backdateTask(t *testing.T, taskID int64) time.Time {
	t.Helper()
	if err := testDB.Model(&models.Task{}).Where("id = ?", taskID).
		UpdateColumns(map[string]any{"updated_at": time.Now().Add(-5 * time.Second)}).Error; err != nil {
		t.Fatalf("回拨任务 %d updated_at: %v", taskID, err)
	}
	return taskUpdatedAt(t, taskID)
}

// backdateProject 回拨项目 updated_at 5 秒并返回真实落库值
func backdateProject(t *testing.T, projectID int64) time.Time {
	t.Helper()
	if err := testDB.Model(&models.Project{}).Where("id = ?", projectID).
		UpdateColumns(map[string]any{"updated_at": time.Now().Add(-5 * time.Second)}).Error; err != nil {
		t.Fatalf("回拨项目 %d updated_at: %v", projectID, err)
	}
	return projectUpdatedAt(t, projectID)
}

// assertAdvanced B2 红线断言：计数变更后该行 updated_at 必须严格前进
func assertAdvanced(t *testing.T, label string, before, after time.Time) {
	t.Helper()
	if !after.After(before) {
		t.Fatalf("B2 违规（%s）：updated_at 未严格前进（before %v, after %v）", label, before, after)
	}
}

// U-S1/E3+E4：任务创建 ⇒ 项目 +1；子任务创建 ⇒ 父直接子 +1；跨项目独立计数
func TestCount_CreateTask_ProjectAndSubtaskIncrement(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	insertProject(t, testProjB)

	parentID := createTask(t, s, taskReq("父任务", testProjA, 0, nil))
	if got := projectTaskCount(t, testProjA); got != 1 {
		t.Fatalf("创建父任务后项目A task_count = %d, want 1", got)
	}
	if _, _, got := taskCountsOf(t, parentID); got != 0 {
		t.Fatalf("父任务自身 subtask_count = %d, want 0", got)
	}

	createTask(t, s, taskReq("子任务", testProjA, parentID, nil))
	if got := projectTaskCount(t, testProjA); got != 2 {
		t.Fatalf("创建子任务后项目A task_count = %d, want 2", got)
	}
	if _, _, got := taskCountsOf(t, parentID); got != 1 {
		t.Fatalf("父任务 subtask_count = %d, want 1", got)
	}

	createTask(t, s, taskReq("项目B任务", testProjB, 0, nil))
	if got := projectTaskCount(t, testProjB); got != 1 {
		t.Fatalf("项目B task_count = %d, want 1", got)
	}
	if got := projectTaskCount(t, testProjA); got != 2 {
		t.Fatalf("项目A task_count 应为 2（B 不影响 A），got %d", got)
	}
}

// U-S1/E1+E2 + U-S3 created=false：检查项/评论 ±1；upsert 覆盖（created=false）不重复 +1
func TestCount_CheckItemCommentLifecycle(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	ensureTestUser(t)
	taskID := createTask(t, s, taskReq("T", testProjA, 0, nil))

	// 创建检查项 ⇒ check_item_count 1
	ciRes, err := s.taskApp.CreateTaskCheckItem(context.Background(), testUserID, &taskDto.CreateTaskCheckItemReq{
		TaskId: idutil.FormatID(taskID),
		Name:   "ci1",
	})
	if err != nil {
		t.Fatalf("CreateTaskCheckItem: %v", err)
	}
	ci, _, _ := taskCountsOf(t, taskID)
	if ci != 1 {
		t.Fatalf("创建检查项后 check_item_count = %d, want 1", ci)
	}
	// U-S3：同 id 覆盖（created=false）不重复 +1
	_, err = s.taskApp.CreateTaskCheckItem(context.Background(), testUserID, &taskDto.CreateTaskCheckItemReq{
		TaskId: idutil.FormatID(taskID),
		Name:   "ci1",
		Id:     &ciRes.Id,
	})
	if err != nil {
		t.Fatalf("覆盖检查项: %v", err)
	}
	ci, _, _ = taskCountsOf(t, taskID)
	if ci != 1 {
		t.Fatalf("覆盖检查项后 check_item_count = %d, want 1（created=false 不重复 +1）", ci)
	}

	// 创建评论 ⇒ comment_count 1
	cmRes, err := s.taskApp.CreateTaskComment(context.Background(), testUserID, &taskDto.CreateTaskCommentReq{
		TaskId:  idutil.FormatID(taskID),
		Content: "c1",
	})
	if err != nil {
		t.Fatalf("CreateTaskComment: %v", err)
	}
	_, cc, _ := taskCountsOf(t, taskID)
	if cc != 1 {
		t.Fatalf("创建评论后 comment_count = %d, want 1", cc)
	}
	// 同 id 覆盖（created=false）不重复 +1
	_, err = s.taskApp.CreateTaskComment(context.Background(), testUserID, &taskDto.CreateTaskCommentReq{
		TaskId:  idutil.FormatID(taskID),
		Content: "c1",
		Id:      &cmRes.Id,
	})
	if err != nil {
		t.Fatalf("覆盖评论: %v", err)
	}
	_, cc, _ = taskCountsOf(t, taskID)
	if cc != 1 {
		t.Fatalf("覆盖评论后 comment_count = %d, want 1", cc)
	}

	// 删除检查项/评论 ⇒ 各自 -1
	if err := s.taskApp.DeleteTaskCheckItem(context.Background(), testUserID, ciRes.Id); err != nil {
		t.Fatalf("DeleteTaskCheckItem: %v", err)
	}
	ci, _, _ = taskCountsOf(t, taskID)
	if ci != 0 {
		t.Fatalf("删除检查项后 check_item_count = %d, want 0", ci)
	}
	if err := s.taskApp.DeleteTaskComment(context.Background(), testUserID, cmRes.Id); err != nil {
		t.Fatalf("DeleteTaskComment: %v", err)
	}
	_, cc, _ = taskCountsOf(t, taskID)
	if cc != 0 {
		t.Fatalf("删除评论后 comment_count = %d, want 0", cc)
	}
}

// U-S2/B2（TC-STAT-S32）：检查项增 / 删两次计数变更后该行 updated_at 均须前进（可被增量拉取发现）
func TestCount_B2_UpdatedAtAdvances(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	ctx := context.Background()
	taskID := createTask(t, s, taskReq("T", testProjA, 0, nil))

	// 检查项增
	baseline := backdateTask(t, taskID)
	ci, err := s.taskApp.CreateTaskCheckItem(ctx, testUserID, &taskDto.CreateTaskCheckItemReq{
		TaskId: idutil.FormatID(taskID),
		Name:   "ci",
	})
	if err != nil {
		t.Fatalf("CreateTaskCheckItem: %v", err)
	}
	assertAdvanced(t, "检查项增", baseline, taskUpdatedAt(t, taskID))

	// 检查项删
	baseline = backdateTask(t, taskID)
	if err := s.taskApp.DeleteTaskCheckItem(ctx, testUserID, ci.Id); err != nil {
		t.Fatalf("DeleteTaskCheckItem: %v", err)
	}
	assertAdvanced(t, "检查项删", baseline, taskUpdatedAt(t, taskID))
}

// TC-STAT-S33~S36 / AC9（B2 红线）：**每条**计数写路径均须 bump 对应行 updated_at
// 覆盖：评论增/删（任务行）、子任务增/换父/脱离（涉及的全部父行）、任务 create/delete/restore/copy（项目行）、
// move（旧/新两个项目行）。
func TestCount_B2_UpdatedAtAdvances_AllWritePaths(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	insertProject(t, testProjB)
	ensureTestUser(t)
	ctx := context.Background()

	parentP := createTask(t, s, taskReq("父P", testProjA, 0, nil))
	parentQ := createTask(t, s, taskReq("父Q", testProjA, 0, nil))
	taskID := createTask(t, s, taskReq("T", testProjA, parentP, nil))

	// S33：评论增 / 删 ⇒ 任务行 bump
	base := backdateTask(t, taskID)
	comment, err := s.taskApp.CreateTaskComment(ctx, testUserID, &taskDto.CreateTaskCommentReq{
		TaskId:  idutil.FormatID(taskID),
		Content: "c1",
	})
	if err != nil {
		t.Fatalf("CreateTaskComment: %v", err)
	}
	assertAdvanced(t, "评论增", base, taskUpdatedAt(t, taskID))

	base = backdateTask(t, taskID)
	if err := s.taskApp.DeleteTaskComment(ctx, testUserID, comment.Id); err != nil {
		t.Fatalf("DeleteTaskComment: %v", err)
	}
	assertAdvanced(t, "评论删", base, taskUpdatedAt(t, taskID))

	// S34：子任务增（父 P）/ 换父（旧父 P + 新父 Q）/ 脱离（旧父 Q）⇒ 涉及的父行均 bump
	baseP := backdateTask(t, parentP)
	childID := createTask(t, s, taskReq("C", testProjA, parentP, nil))
	assertAdvanced(t, "子任务增-父P", baseP, taskUpdatedAt(t, parentP))

	baseP = backdateTask(t, parentP)
	baseQ := backdateTask(t, parentQ)
	if err := s.taskApp.UpdateTask(ctx, testUserID, idutil.FormatID(childID), &taskDto.UpdateTaskReq{
		ParentTaskId: strPtr(idutil.FormatID(parentQ)),
	}); err != nil {
		t.Fatalf("UpdateTask 换父: %v", err)
	}
	assertAdvanced(t, "换父-旧父P", baseP, taskUpdatedAt(t, parentP))
	assertAdvanced(t, "换父-新父Q", baseQ, taskUpdatedAt(t, parentQ))

	baseQ = backdateTask(t, parentQ)
	detach := ""
	if err := s.taskApp.UpdateTask(ctx, testUserID, idutil.FormatID(childID), &taskDto.UpdateTaskReq{
		ParentTaskId: &detach,
	}); err != nil {
		t.Fatalf("UpdateTask 脱离父: %v", err)
	}
	assertAdvanced(t, "脱离-旧父Q", baseQ, taskUpdatedAt(t, parentQ))

	// S35：任务 create / delete / restore / copy ⇒ 项目行 bump
	baseA := backdateProject(t, testProjA)
	t2 := createTask(t, s, taskReq("T2", testProjA, 0, nil))
	assertAdvanced(t, "任务创建-项目A", baseA, projectUpdatedAt(t, testProjA))

	baseA = backdateProject(t, testProjA)
	if err := s.taskApp.DeleteTask(ctx, testUserID, idutil.FormatID(t2)); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	assertAdvanced(t, "任务删除-项目A", baseA, projectUpdatedAt(t, testProjA))

	baseA = backdateProject(t, testProjA)
	if err := s.taskApp.RestoreTask(ctx, testUserID, idutil.FormatID(t2)); err != nil {
		t.Fatalf("RestoreTask: %v", err)
	}
	assertAdvanced(t, "任务恢复-项目A", baseA, projectUpdatedAt(t, testProjA))

	baseA = backdateProject(t, testProjA)
	if _, err := s.taskApp.CopyTask(ctx, testUserID, idutil.FormatID(t2)); err != nil {
		t.Fatalf("CopyTask: %v", err)
	}
	assertAdvanced(t, "任务复制-项目A", baseA, projectUpdatedAt(t, testProjA))

	// S36：move ⇒ 旧/新项目两行均 bump
	baseA = backdateProject(t, testProjA)
	baseB := backdateProject(t, testProjB)
	if err := s.taskApp.UpdateTask(ctx, testUserID, idutil.FormatID(taskID), &taskDto.UpdateTaskReq{
		ProjectId: strPtr(idutil.FormatID(testProjB)),
	}); err != nil {
		t.Fatalf("UpdateTask move: %v", err)
	}
	assertAdvanced(t, "move-旧项目A", baseA, projectUpdatedAt(t, testProjA))
	assertAdvanced(t, "move-新项目B", baseB, projectUpdatedAt(t, testProjB))
}

// TC-STAT-S37 / AC9（B2 的端到端意义）：计数变更后的行必须能被 keyset 增量拉取发现，且载荷为最新计数。
// 游标 = 该行「此前已拉到的版本」；bump 后必须按 strict `>` 被重新拉取（漏拉 = DEF-SYNC-05 式永久陈旧）。
func TestCount_B2_CountChangeIsPullDiscoverable(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	ctx := context.Background()
	repo := NewTaskRepo(testDB)
	taskID := createTask(t, s, taskReq("T", testProjA, 0, nil))

	// 客户端最后一次拉到的版本（已拉过该行）
	cursor := backdateTask(t, taskID)
	if _, err := s.taskApp.CreateTaskCheckItem(ctx, testUserID, &taskDto.CreateTaskCheckItemReq{
		TaskId: idutil.FormatID(taskID),
		Name:   "ci",
	}); err != nil {
		t.Fatalf("CreateTaskCheckItem: %v", err)
	}

	page, err := repo.ListSync(ctx, testUserID, cursor, taskID, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	var pulled *entities.Task
	for _, e := range page {
		if e.Id == taskID {
			pulled = e
		}
	}
	if pulled == nil {
		t.Fatalf("B2 违规：计数变更后该行未被增量拉取发现（cursor=%v, 返回 %d 条）", cursor, len(page))
	}
	if pulled.CheckItemCount != 1 {
		t.Fatalf("增量拉取到的 check_item_count = %d, want 1（载荷须为最新计数）", pulled.CheckItemCount)
	}
}

// TC-STAT-S42 / AC10（B3 server-owned 负向）：请求体携带计数字段 ⇒ 被忽略且不落库。
// 覆盖 CreateTaskReq 与 UpdateTaskReq（JSON 绑定目标），计数只能由服务端事件联动决定。
func TestCount_ServerOwned_RequestCountsIgnored(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	ensureTestUser(t)
	ctx := context.Background()

	// 1. CreateTask 请求体伪造计数 ⇒ 不落库（服务端决定为 0）
	createBody := fmt.Sprintf(`{"name":"伪造计数","state":"todo","priority":"medium","projectId":%q,`+
		`"checkItemCount":99,"commentCount":88,"subtaskCount":77}`, idutil.FormatID(testProjA))
	var createReq taskDto.CreateTaskReq
	if err := json.Unmarshal([]byte(createBody), &createReq); err != nil {
		t.Fatalf("解析伪造 CreateTaskReq: %v", err)
	}
	taskID := createTask(t, s, &createReq)
	if ci, cc, sc := taskCountsOf(t, taskID); ci != 0 || cc != 0 || sc != 0 {
		t.Fatalf("CreateTask 计数被客户端注入：got %d/%d/%d, want 0/0/0", ci, cc, sc)
	}

	// 前置：真实产生 1 个检查项（服务端计数 = 1）
	if _, err := s.taskApp.CreateTaskCheckItem(ctx, testUserID, &taskDto.CreateTaskCheckItemReq{
		TaskId: idutil.FormatID(taskID),
		Name:   "ci",
	}); err != nil {
		t.Fatalf("CreateTaskCheckItem: %v", err)
	}

	// 2. UpdateTask 请求体伪造计数 ⇒ 计数不被覆盖（仍 1/0/0），同请求其他字段正常生效
	var updateReq taskDto.UpdateTaskReq
	if err := json.Unmarshal([]byte(`{"name":"改名","checkItemCount":99,"commentCount":88,"subtaskCount":77}`), &updateReq); err != nil {
		t.Fatalf("解析伪造 UpdateTaskReq: %v", err)
	}
	if err := s.taskApp.UpdateTask(ctx, testUserID, idutil.FormatID(taskID), &updateReq); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if ci, cc, sc := taskCountsOf(t, taskID); ci != 1 || cc != 0 || sc != 0 {
		t.Fatalf("UpdateTask 计数被客户端覆盖：got %d/%d/%d, want 1/0/0", ci, cc, sc)
	}
	var m models.Task
	if err := testDB.Where("id = ?", taskID).First(&m).Error; err != nil {
		t.Fatalf("读任务: %v", err)
	}
	if m.Name != "改名" {
		t.Fatalf("同请求其他字段应正常生效：name = %q, want 改名", m.Name)
	}
}

// U-S3/LWW：更新请求 updatedAt 过期 ⇒ 更新被拒 ⇒ 不发布事件、计数不变
func TestCount_LWWRejected_NoPublish(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	insertProject(t, testProjB)
	taskID := createTask(t, s, taskReq("T", testProjA, 0, nil))

	stale := "2000-01-01T00:00:00Z"
	err := s.taskApp.UpdateTask(context.Background(), testUserID, idutil.FormatID(taskID), &taskDto.UpdateTaskReq{
		ProjectId: strPtr(idutil.FormatID(testProjB)),
		UpdatedAt: &stale,
	})
	if err != nil {
		t.Fatalf("UpdateTask(过期 LWW): %v", err)
	}
	if got := projectTaskCount(t, testProjA); got != 1 {
		t.Fatalf("LWW 拒绝后项目A task_count = %d, want 1", got)
	}
	if got := projectTaskCount(t, testProjB); got != 0 {
		t.Fatalf("LWW 拒绝后项目B task_count = %d, want 0（不应发布 E5）", got)
	}
}

// U-S3/B6：墓碑复活（upsert 覆盖已软删记录）⇒ created=true ⇒ 重新 +1
func TestCount_TombstoneRevive_CountsAsCreated(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	taskID := createTask(t, s, taskReq("T", testProjA, 0, nil))

	if err := s.taskApp.DeleteTask(context.Background(), testUserID, idutil.FormatID(taskID)); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if got := projectTaskCount(t, testProjA); got != 0 {
		t.Fatalf("删除后项目A task_count = %d, want 0", got)
	}

	// 同 id 复活（客户端重放推送）：B6 ⇒ +1
	idStr := idutil.FormatID(taskID)
	createTask(t, s, taskReq("T", testProjA, 0, &idStr))
	if got := projectTaskCount(t, testProjA); got != 1 {
		t.Fatalf("墓碑复活后项目A task_count = %d, want 1", got)
	}
}

// U-S4/E5+E6：move 两项目各 ±1；换父两父各 ±1；脱离父（A→0）仅旧父 -1
func TestCount_MoveAndReparent(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	insertProject(t, testProjB)
	parentP := createTask(t, s, taskReq("父P", testProjA, 0, nil))
	parentQ := createTask(t, s, taskReq("父Q", testProjB, 0, nil))
	taskID := createTask(t, s, taskReq("T", testProjA, parentP, nil))
	if _, _, got := taskCountsOf(t, parentP); got != 1 {
		t.Fatalf("初始父P subtask_count = %d, want 1", got)
	}

	// E5：move A→B（两项目各 ±1）
	if err := s.taskApp.UpdateTask(context.Background(), testUserID, idutil.FormatID(taskID), &taskDto.UpdateTaskReq{
		ProjectId: strPtr(idutil.FormatID(testProjB)),
	}); err != nil {
		t.Fatalf("UpdateTask move: %v", err)
	}
	if got := projectTaskCount(t, testProjA); got != 1 { // P 仍在 A
		t.Fatalf("move 后项目A task_count = %d, want 1", got)
	}
	if got := projectTaskCount(t, testProjB); got != 2 { // Q + T
		t.Fatalf("move 后项目B task_count = %d, want 2", got)
	}

	// E6：换父 P→Q（两父各 ±1）
	if err := s.taskApp.UpdateTask(context.Background(), testUserID, idutil.FormatID(taskID), &taskDto.UpdateTaskReq{
		ParentTaskId: strPtr(idutil.FormatID(parentQ)),
	}); err != nil {
		t.Fatalf("UpdateTask 换父: %v", err)
	}
	if _, _, got := taskCountsOf(t, parentP); got != 0 {
		t.Fatalf("换父后父P subtask_count = %d, want 0", got)
	}
	if _, _, got := taskCountsOf(t, parentQ); got != 1 {
		t.Fatalf("换父后父Q subtask_count = %d, want 1", got)
	}

	// E6：脱离父（A→0）仅旧父 -1
	clear := ""
	if err := s.taskApp.UpdateTask(context.Background(), testUserID, idutil.FormatID(taskID), &taskDto.UpdateTaskReq{
		ParentTaskId: &clear,
	}); err != nil {
		t.Fatalf("UpdateTask 脱离父: %v", err)
	}
	if _, _, got := taskCountsOf(t, parentQ); got != 0 {
		t.Fatalf("脱离后父Q subtask_count = %d, want 0", got)
	}
}

// U-S4/B8：CopyTask 双事件 —— 父 +1 与项目 +1
func TestCount_CopyTask_DualEvents(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	parentID := createTask(t, s, taskReq("父", testProjA, 0, nil))
	taskID := createTask(t, s, taskReq("T", testProjA, parentID, nil))
	if got := projectTaskCount(t, testProjA); got != 2 {
		t.Fatalf("复制前项目A task_count = %d, want 2", got)
	}
	if _, _, got := taskCountsOf(t, parentID); got != 1 {
		t.Fatalf("复制前父 subtask_count = %d, want 1", got)
	}

	if _, err := s.taskApp.CopyTask(context.Background(), testUserID, idutil.FormatID(taskID)); err != nil {
		t.Fatalf("CopyTask: %v", err)
	}
	if got := projectTaskCount(t, testProjA); got != 3 {
		t.Fatalf("复制后项目A task_count = %d, want 3（E4 +1）", got)
	}
	if _, _, got := taskCountsOf(t, parentID); got != 2 {
		t.Fatalf("复制后父 subtask_count = %d, want 2（E3 +1）", got)
	}
}

// U-S1/E7：项目级联删/恢复 ⇒ 批量重算（写最终值，不逐事件）
func TestCount_ProjectCascadeDeleteRestore_Recount(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	for i := 0; i < 3; i++ {
		createTask(t, s, taskReq("t"+strconv.Itoa(i), testProjA, 0, nil))
	}
	if got := projectTaskCount(t, testProjA); got != 3 {
		t.Fatalf("删除前项目A task_count = %d, want 3", got)
	}

	if err := s.projApp.Delete(context.Background(), testUserID, idutil.FormatID(testProjA)); err != nil {
		t.Fatalf("Project Delete: %v", err)
	}
	if got := projectTaskCount(t, testProjA); got != 0 {
		t.Fatalf("级联删后项目A task_count = %d, want 0（E7 重算）", got)
	}

	if err := s.projApp.Restore(context.Background(), testUserID, idutil.FormatID(testProjA)); err != nil {
		t.Fatalf("Project Restore: %v", err)
	}
	if got := projectTaskCount(t, testProjA); got != 3 {
		t.Fatalf("级联恢复后项目A task_count = %d, want 3（E7 重算）", got)
	}
}

// 集成原子性：计数更新失败 ⇒ 主写整体回滚（计数与主写永不分离）
func TestCount_Atomicity_RollbackOnCountFailure(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)
	// 追加失败订阅者（在真实 updater 之后）：任何计数事件都让事务失败
	s.bus.Subscribe(func(ctx context.Context, event types.CountEvent) error {
		return errors.New("模拟计数更新失败")
	})

	_, err := s.taskApp.CreateTask(context.Background(), testUserID, taskReq("T", testProjA, 0, nil))
	if err == nil {
		t.Fatal("计数更新失败时 CreateTask 应返回错误（事务回滚）")
	}
	var taskCnt int64
	if err := testDB.Model(&models.Task{}).Where("user_id = ?", testUserID).Count(&taskCnt).Error; err != nil {
		t.Fatalf("计数任务行: %v", err)
	}
	if taskCnt != 0 {
		t.Fatalf("原子性违规：主写未回滚，tasks 存在 %d 行", taskCnt)
	}
	if got := projectTaskCount(t, testProjA); got != 0 {
		t.Fatalf("原子性违规：项目A task_count = %d, want 0（计数更新同事务回滚）", got)
	}
}

func strPtr(s string) *string { return &s }
