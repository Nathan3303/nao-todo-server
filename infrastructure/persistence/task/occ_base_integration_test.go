//go:build integration

// T163 契约集成：OCC base（baseUpdatedAt）判定在真实 MySQL 上的落库语义。
//
// 覆盖 ADR §9.1.3 / §9.6 R-10/R-11/R-12：
//   - base 缺失 ⇒ 现行 LWW 零变化（noop / overwrite）
//   - base 与库中版本相等 ⇒ applied（覆盖写入）
//   - base 与库中版本不等 ⇒ stale（不写入 + 回传库中当前版本）
//   - createdAt 相差过大 ⇒ conflict（ID 碰撞语义不变，优先于 base 分支）
package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	apptask "naotodoserver/application/task"
	"naotodoserver/application/task/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/types"
)

// occPushReq 构造一次 sync push 任务条目（含 OCC base）
func occPushReq(id, name, createdStr string, updated time.Time, base *string) *dto.CreateTaskReq {
	updatedStr := updated.Format(time.RFC3339)
	return &dto.CreateTaskReq{
		Id:            &id,
		Name:          name,
		State:         "pending",
		Priority:      "medium",
		CreatedAt:     &createdStr,
		UpdatedAt:     &updatedStr,
		BaseUpdatedAt: base,
	}
}

// TestUpsertOCCBase_Integration base 分支 + LWW 回退 + ID 碰撞优先级
func TestUpsertOCCBase_Integration(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID = 1950
	const idStr = "9901"

	now := time.Now().UTC().Truncate(time.Millisecond)
	createdStr := now.Format(time.RFC3339)
	strPtr := func(s string) *string { return &s }

	// upsert 走 sync push 同款转换链（DTO → VO → repo.Upsert），返回回传实体名 + 判定
	upsert := func(name string, updated time.Time, base *string) (string, types.UpsertResult, error) {
		t.Helper()
		vo, err := apptask.CreateTaskReqToValueObject(userID, occPushReq(idStr, name, createdStr, updated, base))
		if err != nil {
			t.Fatalf("CreateTaskReqToValueObject: %v", err)
		}
		entity, res, err := repo.Upsert(ctx, userID, vo)
		if entity == nil {
			return "", res, err
		}
		return entity.Name, res, err
	}
	readName := func() string {
		t.Helper()
		entity, err := repo.GetById(ctx, userID, 9901, true)
		if err != nil {
			t.Fatalf("GetById: %v", err)
		}
		return entity.Name
	}
	readUpdatedAt := func() time.Time {
		t.Helper()
		entity, err := repo.GetById(ctx, userID, 9901, true)
		if err != nil {
			t.Fatalf("GetById: %v", err)
		}
		return entity.UpdatedAt
	}

	// ① 行不存在 + 携带 base ⇒ 新建 applied（base 不参与判定）
	if _, res, err := upsert("v0", now, strPtr(now.Format(time.RFC3339))); err != nil ||
		res.Outcome != types.UpsertOverwrite || !res.Created {
		t.Fatalf("新建：outcome=%v created=%v err=%v, want Overwrite/created", res.Outcome, res.Created, err)
	}
	serverV1 := readUpdatedAt()

	// ② base == 库中版本 ⇒ applied（覆盖写入）
	time.Sleep(5 * time.Millisecond)
	base1 := serverV1.Format(idutil.RFC3339Milli)
	if _, res, err := upsert("v1", now.Add(time.Second), &base1); err != nil ||
		res.Outcome != types.UpsertOverwrite {
		t.Fatalf("base 匹配：outcome=%v err=%v, want Overwrite", res.Outcome, err)
	}
	if got := readName(); got != "v1" {
		t.Fatalf("base 匹配应写入，DB name=%q want v1", got)
	}
	serverV2 := readUpdatedAt()
	if !serverV2.After(serverV1) {
		t.Fatalf("覆盖后服务端版本应推进: %v -> %v", serverV1, serverV2)
	}

	// ③ base != 库中版本 ⇒ stale：不写入 + 回传库中当前版本
	time.Sleep(5 * time.Millisecond)
	staleBase := base1 // 陈旧 base（serverV1）
	returnedName, res, err := upsert("v2-should-not-write", now.Add(2*time.Second), &staleBase)
	if err != nil || res.Outcome != types.UpsertStale {
		t.Fatalf("base 不匹配：outcome=%v err=%v, want Stale", res.Outcome, err)
	}
	if got := readName(); got != "v1" {
		t.Fatalf("stale 不应写入，DB name=%q want v1", got)
	}
	if got := readUpdatedAt(); !got.Equal(serverV2) {
		t.Fatalf("stale 后库中版本不应变化: %v -> %v", serverV2, got)
	}
	if returnedName != "v1" {
		t.Fatalf("stale 应回传库中当前实体（供 rebase）: name=%q want v1", returnedName)
	}

	// ④ base 缺失（nil）⇒ 现行 LWW：请求更旧 → noop（零行为变化）
	time.Sleep(5 * time.Millisecond)
	if _, res, err := upsert("v3", now.Add(-time.Hour), nil); err != nil ||
		res.Outcome != types.UpsertNoop {
		t.Fatalf("base 缺失 + 请求更旧：outcome=%v err=%v, want Noop", res.Outcome, err)
	}
	if got := readName(); got != "v1" {
		t.Fatalf("noop 不应写入，DB name=%q want v1", got)
	}

	// ⑤ createdAt 相差过大 ⇒ conflict（ID 碰撞优先于 base 分支）
	farCreated := now.Add(-2 * time.Hour).Format(time.RFC3339)
	vo, err := apptask.CreateTaskReqToValueObject(userID, &dto.CreateTaskReq{
		Id:            strPtr(idStr),
		Name:          "collision",
		State:         "pending",
		Priority:      "medium",
		CreatedAt:     &farCreated,
		UpdatedAt:     strPtr(now.Add(3 * time.Second).Format(time.RFC3339)),
		BaseUpdatedAt: &staleBase,
	})
	if err != nil {
		t.Fatalf("CreateTaskReqToValueObject: %v", err)
	}
	// ID 碰撞以 error 表达（既有语义，控制器据 error → outcome=conflict）
	if _, _, err := repo.Upsert(ctx, userID, vo); !errors.Is(err, domerr.ErrIDConflict) {
		t.Fatalf("ID 碰撞：err=%v, want ErrIDConflict", err)
	}
	if got := readName(); got != "v1" {
		t.Fatalf("conflict 不应写入，DB name=%q want v1", got)
	}
}
