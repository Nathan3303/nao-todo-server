//go:build integration

// 子任务排序（sort_id）—— 全链路集成测试（真实 MySQL + 完整装配：app + 内存总线 + CountUpdater）
// 覆盖 ADR 2026-09-13 §10 服务端口径：U-S1 组内 max、U-S2 G1–G5、U-S3 G6–G8、U-S4 G9/G10、
// U-S5 BC（默认序 + 显式 sort 优先）、U-S6 组隔离 + 计数零变化（AC9/AC15）。
// 依赖 repoImpl_integration_test.go 的 TestMain 与 count_events_integration_test.go 的装配/辅助函数。
package task

import (
	"context"
	"errors"
	"testing"

	"naotodoserver/application/idutil"
	taskDto "naotodoserver/application/task/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/infrastructure/persistence/models"
)

// u16ptr uint16 指针（PATCH 显式排序值）
func u16ptr(v uint16) *uint16 { return &v }

// taskSortId 读任务 sort_id 列
func taskSortId(t *testing.T, taskID int64) uint16 {
	t.Helper()
	var m models.Task
	if err := testDB.Where("id = ?", taskID).First(&m).Error; err != nil {
		t.Fatalf("读任务 %d: %v", taskID, err)
	}
	return m.SortId
}

// taskName 读任务名称
func taskName(t *testing.T, taskID int64) string {
	t.Helper()
	var m models.Task
	if err := testDB.Where("id = ?", taskID).First(&m).Error; err != nil {
		t.Fatalf("读任务 %d: %v", taskID, err)
	}
	return m.Name
}

// updateTaskViaApp 经 app 层 PATCH 任务
func updateTaskViaApp(t *testing.T, s *fullStack, taskID int64, req *taskDto.UpdateTaskReq) error {
	t.Helper()
	return s.taskApp.UpdateTask(context.Background(), testUserID, idutil.FormatID(taskID), req)
}

// deleteTaskViaApp 经 app 层软删任务（墓碑）
func deleteTaskViaApp(t *testing.T, s *fullStack, taskID int64) {
	t.Helper()
	if err := s.taskApp.DeleteTask(context.Background(), testUserID, idutil.FormatID(taskID)); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
}

// copyTaskViaApp 经 app 层复制任务，返回复制品 ID
func copyTaskViaApp(t *testing.T, s *fullStack, taskID int64) int64 {
	t.Helper()
	res, err := s.taskApp.CopyTask(context.Background(), testUserID, idutil.FormatID(taskID))
	if err != nil {
		t.Fatalf("CopyTask: %v", err)
	}
	id, err := idutil.ParseID(res.Id)
	if err != nil {
		t.Fatalf("解析复制品 ID %q: %v", res.Id, err)
	}
	return id
}

// listSubTasks 按组查询（parentTaskId 过滤）任务列表
func listSubTasks(
	t *testing.T,
	s *fullStack,
	parentTaskID int64,
	sort string,
) (taskDto.ListTaskRes, *taskDto.Pagination) {
	t.Helper()
	res, pagination, err := s.taskApp.ListTask(context.Background(), testUserID, &taskDto.ListTaskReq{
		ParentTaskId: idutil.FormatID(parentTaskID),
		Limit:        50,
		Sort:         sort,
	})
	if err != nil {
		t.Fatalf("ListTask(parent=%d): %v", parentTaskID, err)
	}
	return res, pagination
}

// resIds 提取列表响应 ID（int64，保持响应顺序）
func resIds(t *testing.T, res taskDto.ListTaskRes) []int64 {
	t.Helper()
	ids := make([]int64, 0, len(res))
	for _, item := range res {
		id, err := idutil.ParseID(item.Id)
		if err != nil {
			t.Fatalf("解析任务 ID %q: %v", item.Id, err)
		}
		ids = append(ids, id)
	}
	return ids
}

// insertLegacyTask 直插存量行（模拟历史版本写入的 sort_id = 0，BC/U-S5 用）
func insertLegacyTask(t *testing.T, name string, parentTaskID int64, sortId uint16) int64 {
	t.Helper()
	m := &models.Task{
		UserId:       testUserID,
		ProjectId:    testProjA,
		ParentTaskId: parentTaskID,
		Name:         name,
		State:        1,
		Priority:     2,
		SortId:       sortId,
	}
	if err := testDB.Create(m).Error; err != nil {
		t.Fatalf("直插存量任务 %q: %v", name, err)
	}
	return m.ID
}

// reqWithSort 构造带显式 sortId 的创建请求
func reqWithSort(name string, projectID, parentTaskID int64, sortId uint16, id *string) *taskDto.CreateTaskReq {
	req := taskReq(name, projectID, parentTaskID, id)
	req.SortId = sortId
	return req
}

// taskCounts 三计数快照（便于整体比较，AC15 用）
type taskCounts struct {
	checkItems, comments, subtasks uint
}

// countsOf 读取任务三计数快照
func countsOf(t *testing.T, taskID int64) taskCounts {
	t.Helper()
	ci, cm, st := taskCountsOf(t, taskID)
	return taskCounts{checkItems: ci, comments: cm, subtasks: st}
}

// --- U-S1：GetMaxSortId 按组取值（组 0 与子组互不影响；软删不计；空组 = 255 基线 ⇒ 分配 256） ---

func TestSortId_US1_GetMaxSortIdPerGroup(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	ctx := context.Background()
	repo := NewTaskRepo(testDB)

	// 空组基线：255（调用方 +1 = 256）
	for _, group := range []int64{0, 999999} {
		got, err := repo.GetMaxSortId(ctx, testUserID, group)
		if err != nil {
			t.Fatalf("空组 %d GetMaxSortId: %v", group, err)
		}
		if got != 255 {
			t.Fatalf("空组 %d max = %d, want 255（基线）", group, got)
		}
	}

	// 组 0：依次 256、257、258
	top1 := createTask(t, s, taskReq("top1", testProjA, 0, nil))
	top2 := createTask(t, s, taskReq("top2", testProjA, 0, nil))
	parent := createTask(t, s, taskReq("父", testProjA, 0, nil))
	if got := taskSortId(t, top1); got != 256 {
		t.Fatalf("组0 首个 sort_id = %d, want 256", got)
	}
	if got := taskSortId(t, top2); got != 257 {
		t.Fatalf("组0 次个 sort_id = %d, want 257", got)
	}
	if got := taskSortId(t, parent); got != 258 {
		t.Fatalf("组0 三个 sort_id = %d, want 258", got)
	}

	// 子组独立：组 parent 首个 = 256（与组 0 的 258 无关）
	child := createTask(t, s, taskReq("子", testProjA, parent, nil))
	if got := taskSortId(t, child); got != 256 {
		t.Fatalf("子组首个 sort_id = %d, want 256（组内独立）", got)
	}
	if got, err := repo.GetMaxSortId(ctx, testUserID, 0); err != nil || got != 258 {
		t.Fatalf("组0 max = %d (err %v), want 258", got, err)
	}
	if got, err := repo.GetMaxSortId(ctx, testUserID, parent); err != nil || got != 256 {
		t.Fatalf("子组 max = %d (err %v), want 256", got, err)
	}

	// 软删（墓碑）行不计：删掉子组最大者 ⇒ 回到空组基线
	deleteTaskViaApp(t, s, child)
	if got, err := repo.GetMaxSortId(ctx, testUserID, parent); err != nil || got != 255 {
		t.Fatalf("软删后子组 max = %d (err %v), want 255（软删行不计）", got, err)
	}
	// 组 0 不受影响
	if got, err := repo.GetMaxSortId(ctx, testUserID, 0); err != nil || got != 258 {
		t.Fatalf("软删子任务后组0 max = %d (err %v), want 258", got, err)
	}
}

// --- U-S2：G1–G5 创建/覆盖矩阵（显式优先；未带仅新建或换父置新组末；父未变不写列） ---

func TestSortId_US2_CreateMatrix(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	ctx := context.Background()
	repo := NewTaskRepo(testDB)
	insertProject(t, testProjA)

	parentA := createTask(t, s, taskReq("父A", testProjA, 0, nil)) // 组0 #1 = 256
	parentB := createTask(t, s, taskReq("父B", testProjA, 0, nil)) // 组0 #2 = 257

	// G1：新建显式非零 ⇒ 请求值（断言不被领域层 max+1 覆盖，防回归 serviceImpl.go:28）
	g1 := createTask(t, s, reqWithSort("g1", testProjA, 0, 5000, nil))
	if got := taskSortId(t, g1); got != 5000 {
		t.Fatalf("G1 显式 sortId 被覆盖：sort_id = %d, want 5000", got)
	}
	g1sub := createTask(t, s, reqWithSort("g1-子", testProjA, parentA, 4000, nil))
	if got := taskSortId(t, g1sub); got != 4000 {
		t.Fatalf("G1(子组) 显式 sortId 被覆盖：sort_id = %d, want 4000", got)
	}

	// G2：新建未带 ⇒ 目标组组末（组 0 max 5000 ⇒ 5001）
	g2 := createTask(t, s, taskReq("g2", testProjA, 0, nil))
	if got := taskSortId(t, g2); got != 5001 {
		t.Fatalf("G2 组末：组0 sort_id = %d, want 5001", got)
	}
	// G2（子组）：父A 组 max 4000 ⇒ 4001、4002
	c1 := createTask(t, s, taskReq("c1", testProjA, parentA, nil))
	c2 := createTask(t, s, taskReq("c2", testProjA, parentA, nil))
	if got := taskSortId(t, c1); got != 4001 {
		t.Fatalf("G2 子组首个：sort_id = %d, want 4001", got)
	}
	if got := taskSortId(t, c2); got != 4002 {
		t.Fatalf("G2 子组次个：sort_id = %d, want 4002", got)
	}

	// G2'（ADR 矩阵按 Id==0 表述，实现按「行将被创建」判定）：push 带客户端 id 的新建 ⇒ 组末
	pushNewId := idutil.FormatID(int64(777000111))
	pushNew := createTask(t, s, taskReq("push-new", testProjA, parentA, &pushNewId))
	if got := taskSortId(t, pushNew); got != 4003 {
		t.Fatalf("G2'(push 新建带 id) 组末：sort_id = %d, want 4003", got)
	}

	// G3：覆盖 + 显式非零 ⇒ 请求值
	c1Id := idutil.FormatID(c1)
	createTask(t, s, reqWithSort("c1-覆盖G3", testProjA, parentA, 6000, &c1Id))
	if got := taskSortId(t, c1); got != 6000 {
		t.Fatalf("G3 覆盖显式：sort_id = %d, want 6000", got)
	}

	// G4：覆盖 + 未带 + 父未变 ⇒ 不写列（B1：不得清零组内序）
	createTask(t, s, taskReq("c1-覆盖G4", testProjA, parentA, &c1Id))
	if got := taskSortId(t, c1); got != 6000 {
		t.Fatalf("G4 父未变不得写列：sort_id = %d, want 6000（保持）", got)
	}
	if got := taskName(t, c1); got != "c1-覆盖G4" {
		t.Fatalf("G4 覆盖未生效：name = %q, want c1-覆盖G4", got)
	}

	// G5：覆盖 + 未带 + 换父 ⇒ 新组组末（父B 空组 ⇒ 256）；旧组其它行不动
	createTask(t, s, taskReq("c1-换父G5", testProjA, parentB, &c1Id))
	if got := taskSortId(t, c1); got != 256 {
		t.Fatalf("G5 换父置新组末：sort_id = %d, want 256", got)
	}
	if got, err := repo.GetMaxSortId(ctx, testUserID, parentB); err != nil || got != 256 {
		t.Fatalf("G5 后父B组 max = %d (err %v), want 256", got, err)
	}
	// 旧组（父A）其余行 sort_id 不变
	for _, tc := range []struct {
		name string
		id   int64
		want uint16
	}{
		{"g1-子", g1sub, 4000},
		{"c2", c2, 4002},
		{"push-new", pushNew, 4003},
	} {
		if got := taskSortId(t, tc.id); got != tc.want {
			t.Fatalf("G5 后旧组 %s sort_id = %d, want %d（旧组不动）", tc.name, got, tc.want)
		}
	}
}

// --- U-S3：G6–G8 PATCH 矩阵 ---

func TestSortId_US3_UpdateMatrix(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	ctx := context.Background()
	repo := NewTaskRepo(testDB)
	insertProject(t, testProjA)

	parentA := createTask(t, s, taskReq("父A", testProjA, 0, nil)) // 组0 = 256
	parentB := createTask(t, s, taskReq("父B", testProjA, 0, nil)) // 组0 = 257
	child := createTask(t, s, taskReq("子", testProjA, parentA, nil))

	// G6：PATCH 显式 sortId ⇒ 请求值
	if err := updateTaskViaApp(t, s, child, &taskDto.UpdateTaskReq{SortId: u16ptr(1234)}); err != nil {
		t.Fatalf("G6 UpdateTask: %v", err)
	}
	if got := taskSortId(t, child); got != 1234 {
		t.Fatalf("G6 显式 PATCH：sort_id = %d, want 1234", got)
	}

	// G7：换父未带 sortId ⇒ 新组组末（父B 空组 ⇒ 256）
	if err := updateTaskViaApp(t, s, child, &taskDto.UpdateTaskReq{
		ParentTaskId: strPtr(idutil.FormatID(parentB)),
	}); err != nil {
		t.Fatalf("G7 UpdateTask: %v", err)
	}
	if got := taskSortId(t, child); got != 256 {
		t.Fatalf("G7 换父置新组末：sort_id = %d, want 256", got)
	}
	// G7 边界：脱离父（→ 组 0）⇒ 组 0 组末（组 0 max 257 ⇒ 258）
	if err := updateTaskViaApp(t, s, child, &taskDto.UpdateTaskReq{
		ParentTaskId: strPtr(""),
	}); err != nil {
		t.Fatalf("G7(脱离) UpdateTask: %v", err)
	}
	if got := taskSortId(t, child); got != 258 {
		t.Fatalf("G7 脱离置组0末：sort_id = %d, want 258", got)
	}

	// G8：父未变（重复提交同一 parentTaskId）且未带 sortId ⇒ 不写列、不重排
	if err := updateTaskViaApp(t, s, child, &taskDto.UpdateTaskReq{
		ParentTaskId: strPtr(""),
	}); err != nil {
		t.Fatalf("G8 UpdateTask: %v", err)
	}
	if got := taskSortId(t, child); got != 258 {
		t.Fatalf("G8 父未变不得重排：sort_id = %d, want 258（保持）", got)
	}
	// G8'：纯字段 PATCH（无父/无 sortId）⇒ 不写列
	if err := updateTaskViaApp(t, s, child, &taskDto.UpdateTaskReq{Name: strPtr("子-改名")}); err != nil {
		t.Fatalf("G8' UpdateTask: %v", err)
	}
	if got := taskSortId(t, child); got != 258 {
		t.Fatalf("G8' 纯字段 PATCH：sort_id = %d, want 258（保持）", got)
	}

	// AC15/B8：纯重排不触发计数事件，且 bump 自身 updated_at（§6 增量可发现）
	before := backdateTask(t, child)
	countsBeforeP := countsOf(t, parentA)
	projBefore := projectTaskCount(t, testProjA)
	if err := updateTaskViaApp(t, s, child, &taskDto.UpdateTaskReq{SortId: u16ptr(999)}); err != nil {
		t.Fatalf("纯重排 UpdateTask: %v", err)
	}
	if got := taskSortId(t, child); got != 999 {
		t.Fatalf("纯重排：sort_id = %d, want 999", got)
	}
	assertAdvanced(t, "纯重排 bump 自身 updated_at", before, taskUpdatedAt(t, child))
	if after := countsOf(t, parentA); after != countsBeforeP {
		t.Fatalf("纯重排改变了父计数：before %v, after %v（B8）", countsBeforeP, after)
	}
	if after := projectTaskCount(t, testProjA); after != projBefore {
		t.Fatalf("纯重排改变了项目计数：before %d, after %d（B8）", projBefore, after)
	}
	_ = ctx
	_ = repo
}

// --- U-S4：G9 Copy 置源父组末；G10 Restore 保留原值不重排 ---

func TestSortId_US4_CopyAndRestore(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)

	parent := createTask(t, s, taskReq("父", testProjA, 0, nil))
	createTask(t, s, reqWithSort("组内1", testProjA, parent, 1000, nil))
	createTask(t, s, reqWithSort("组内2", testProjA, parent, 2000, nil))
	src := createTask(t, s, reqWithSort("源", testProjA, parent, 1500, nil))

	// G9：复制品置源父组末（≠ 0，避免落到组首）
	copied := copyTaskViaApp(t, s, src)
	if got := taskSortId(t, copied); got != 2001 {
		t.Fatalf("G9 Copy 置组末：sort_id = %d, want 2001", got)
	}

	// G10：Restore 保留原 sort_id，不重排到组末
	original := taskSortId(t, src) // 1500
	deleteTaskViaApp(t, s, src)
	if err := s.taskApp.RestoreTask(context.Background(), testUserID, idutil.FormatID(src)); err != nil {
		t.Fatalf("RestoreTask: %v", err)
	}
	if got := taskSortId(t, src); got != original {
		t.Fatalf("G10 Restore 不应重排：sort_id = %d, want %d（保留原值）", got, original)
	}
}

// --- U-S5：BC —— 带 parentTaskId 过滤时默认序 sort_id ASC, id ASC；显式 sort 语义不变 ---

func TestSortId_US5_ListDefaultOrder(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)

	parent := createTask(t, s, taskReq("父", testProjA, 0, nil))
	// 显式乱序写入：插入顺序 3000 → 1000 → 2000
	r3 := createTask(t, s, reqWithSort("r3", testProjA, parent, 3000, nil))
	r1 := createTask(t, s, reqWithSort("r1", testProjA, parent, 1000, nil))
	r2 := createTask(t, s, reqWithSort("r2", testProjA, parent, 2000, nil))
	// 存量行：sort_id = 0（AC10 兜底）+ 同值行（同值由 id 二级键兜底，B9）
	legacyA := insertLegacyTask(t, "legacyA", parent, 0)
	legacyB := insertLegacyTask(t, "legacyB", parent, 0)
	same1 := createTask(t, s, reqWithSort("same1", testProjA, parent, 1500, nil))
	same2 := createTask(t, s, reqWithSort("same2", testProjA, parent, 1500, nil))

	// 默认序：sort_id ASC, id ASC（0 值在最前，且同值/同 0 按 id 升序）
	res, pagination := listSubTasks(t, s, parent, "")
	got := resIds(t, res)
	want := []int64{legacyA, legacyB, r1, same1, same2, r2, r3}
	if len(got) != len(want) {
		t.Fatalf("默认序条数 = %d, want %d（%v）", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("默认序 = %v, want %v（sort_id ASC, id ASC）", got, want)
		}
	}
	// Q2：pagination.total 必须准确（客户端用它判「未取尽」以禁重建）
	if pagination == nil || pagination.Total != int64(len(want)) {
		t.Fatalf("pagination.total = %v, want %d", pagination, len(want))
	}

	// 显式 sort 语义不变：sortId:desc（不做默认序追加）；同值行间无二级键 ⇒ 仅断言 sort_id 序列非递增
	resDesc, _ := listSubTasks(t, s, parent, "sortId:desc")
	wantSortIds := []uint16{3000, 2000, 1500, 1500, 1000, 0, 0}
	if len(resDesc) != len(wantSortIds) {
		t.Fatalf("显式 sortId:desc 条数 = %d, want %d", len(resDesc), len(wantSortIds))
	}
	for i, item := range resDesc {
		if item.SortId != wantSortIds[i] {
			t.Fatalf("显式 sortId:desc sort_id[%d] = %d, want %d（非递增）", i, item.SortId, wantSortIds[i])
		}
	}
}

// --- U-S6：组隔离（AC9）+ 计数零变化（AC15） ---

func TestSortId_US6_GroupIsolationAndCountsUntouched(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)

	groupP := createTask(t, s, taskReq("组P", testProjA, 0, nil))
	groupQ := createTask(t, s, taskReq("组Q", testProjA, 0, nil))
	p1 := createTask(t, s, taskReq("p1", testProjA, groupP, nil))
	p2 := createTask(t, s, taskReq("p2", testProjA, groupP, nil))
	q1 := createTask(t, s, reqWithSort("q1", testProjA, groupQ, 7000, nil))
	q2 := createTask(t, s, reqWithSort("q2", testProjA, groupQ, 8000, nil))

	countsBeforeP := countsOf(t, groupP)
	countsBeforeQ := countsOf(t, groupQ)
	projBefore := projectTaskCount(t, testProjA)

	// 重排组 P（1000,2000）—— 模拟客户端 batchUpdate 逐条写入
	if err := updateTaskViaApp(t, s, p1, &taskDto.UpdateTaskReq{SortId: u16ptr(1000)}); err != nil {
		t.Fatalf("重排 p1: %v", err)
	}
	if err := updateTaskViaApp(t, s, p2, &taskDto.UpdateTaskReq{SortId: u16ptr(2000)}); err != nil {
		t.Fatalf("重排 p2: %v", err)
	}

	// 组 P 序正确
	resP, _ := listSubTasks(t, s, groupP, "")
	gotP := resIds(t, resP)
	if len(gotP) != 2 || gotP[0] != p1 || gotP[1] != p2 {
		t.Fatalf("组P 重排后 = %v, want [%d %d]", gotP, p1, p2)
	}

	// AC9：组 Q 各 sort_id 不变、顺序不变
	if got := taskSortId(t, q1); got != 7000 {
		t.Fatalf("组Q q1 sort_id = %d, want 7000（隔离）", got)
	}
	if got := taskSortId(t, q2); got != 8000 {
		t.Fatalf("组Q q2 sort_id = %d, want 8000（隔离）", got)
	}
	resQ, _ := listSubTasks(t, s, groupQ, "")
	gotQ := resIds(t, resQ)
	if len(gotQ) != 2 || gotQ[0] != q1 || gotQ[1] != q2 {
		t.Fatalf("组Q 顺序 = %v, want [%d %d]（隔离）", gotQ, q1, q2)
	}

	// AC15：计数零变化
	if after := countsOf(t, groupP); after != countsBeforeP {
		t.Fatalf("组P 计数变化：before %v, after %v（AC15）", countsBeforeP, after)
	}
	if after := countsOf(t, groupQ); after != countsBeforeQ {
		t.Fatalf("组Q 计数变化：before %v, after %v（AC15）", countsBeforeQ, after)
	}
	if after := projectTaskCount(t, testProjA); after != projBefore {
		t.Fatalf("项目计数变化：before %d, after %d（AC15）", projBefore, after)
	}
}

// --- Q1：组内排序值耗尽（max = 65535）⇒ 返回可识别的领域错误且不写入 ---

func TestSortId_OverflowReturnsDomainError(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)

	parent := createTask(t, s, taskReq("父", testProjA, 0, nil))
	// 直插组内 max = 65535（模拟组内排序值耗尽）
	insertLegacyTask(t, "满了", parent, 65535)

	// 新建未带 ⇒ max+1 = 65536 溢出 ⇒ 明确领域错误，不写入
	req := taskReq("溢出", testProjA, parent, nil)
	_, err := s.taskApp.CreateTask(context.Background(), testUserID, req)
	if !errors.Is(err, domerr.ErrSortIdOverflow) {
		t.Fatalf("溢出创建 err = %v, want ErrSortIdOverflow", err)
	}
	var cnt int64
	if err := testDB.Model(&models.Task{}).Where("name = ?", "溢出").Count(&cnt).Error; err != nil {
		t.Fatalf("统计溢出行: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("溢出时不得写入：脏行数 = %d, want 0", cnt)
	}

	// PATCH 换父到该组（未带 sortId）⇒ 同样返回领域错误
	other := createTask(t, s, taskReq("游离", testProjA, 0, nil))
	err = updateTaskViaApp(t, s, other, &taskDto.UpdateTaskReq{
		ParentTaskId: strPtr(idutil.FormatID(parent)),
	})
	if !errors.Is(err, domerr.ErrSortIdOverflow) {
		t.Fatalf("溢出换父 err = %v, want ErrSortIdOverflow", err)
	}
	// 换父失败 ⇒ 事务回滚，父未变（游离 = 组0 第二个 = 257）
	if got := taskSortId(t, other); got != 257 {
		t.Fatalf("溢出失败后回归滚：游离 sort_id = %d, want 257", got)
	}
}
