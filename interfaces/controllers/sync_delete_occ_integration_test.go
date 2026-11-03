//go:build integration

// T642 / A4 端到端集成测试（真实 MySQL + 完整装配 + 真实 HTTP 绑定）：
// /sync/push 的 deletions 携带 baseUpdatedAt 时按 OCC 裁决，且回执 serverUpdatedAt 是库中真实值。
//
// 覆盖：base 不匹配 ⇒ stale + serverCurrent（库中未变）；base 命中 ⇒ applied；重复删除 ⇒ 幂等；
// 老客户端（无 base）⇒ 删除优先（R-D 逐字旧行为）。运行：
//
//	TZ=UTC go test -tags integration -p 1 -count=1 -run TestSyncPushDeletion_OCC -v ./interfaces/controllers/
package controllers

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/interfaces/types"
)

const doccUID int64 = 77001

// doccTaskRow 读库中任务行（含墓碑），返回 (updated_at RFC3339Milli, 是否墓碑)
func doccTaskRow(t *testing.T, taskID string) (string, bool) {
	t.Helper()
	id, err := idutil.ParseID(taskID)
	if err != nil {
		t.Fatalf("解析 id %q: %v", taskID, err)
	}
	var m models.Task
	if err := nullClearDB.Unscoped().Where("id = ? AND user_id = ?", id, doccUID).First(&m).Error; err != nil {
		t.Fatalf("读任务 %s: %v", taskID, err)
	}
	return m.UpdatedAt.Format(idutil.RFC3339Milli), m.DeletedAt.Valid
}

// doccServerCurrent 取回执里的 serverCurrent 原始对象（断言字段出现与否）
func doccServerCurrent(t *testing.T, r types.SyncResult) (map[string]any, bool) {
	t.Helper()
	if r.ServerCurrent == nil {
		return nil, false
	}
	b, err := json.Marshal(r.ServerCurrent)
	if err != nil {
		t.Fatalf("marshal serverCurrent: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal serverCurrent: %v", err)
	}
	return m, true
}

// doccDeletePush 发起一次 deletions 推送并返回唯一一条结果
func doccDeletePush(t *testing.T, ctrl *SyncController, body string) types.SyncResult {
	t.Helper()
	resp := doDerivedPush(t, ctrl, doccUID, body)
	if len(resp.Results) != 1 {
		t.Fatalf("results 条数 = %d, want 1", len(resp.Results))
	}
	return resp.Results[0]
}

// TestSyncPushDeletion_OCC_HTTP /sync/push 删除 OCC 端到端（含回传真实 updated_at）
func TestSyncPushDeletion_OCC_HTTP(t *testing.T) {
	ctrl := newDerivedController()
	cleanNullClear(t, doccUID)

	const taskID = "97001"
	// 1. 先推一条任务（新建）
	resp := doDerivedPush(t, ctrl, doccUID,
		fmt.Sprintf(`{"tasks":[{"id":"%s","name":"T","state":"pending","priority":"medium"}]}`, taskID))
	if resp.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("新建任务 outcome = %q, want applied", resp.Results[0].Outcome)
	}
	baseUpdatedAt, deleted := doccTaskRow(t, taskID)
	if deleted {
		t.Fatal("新建任务不应是墓碑")
	}

	// 2. base 不匹配 ⇒ stale，回带 serverCurrent（服务端当前行），库中不变、不产生新墓碑
	staleBase := mustParseDocc(t, baseUpdatedAt).Add(-time.Second).Format(time.RFC3339Nano)
	stale := doccDeletePush(t, ctrl, fmt.Sprintf(
		`{"deletions":[{"table":"tasks","id":"%s","baseUpdatedAt":"%s"}]}`, taskID, staleBase))
	if stale.Outcome != types.SyncOutcomeStale {
		t.Fatalf("outcome = %q, want stale", stale.Outcome)
	}
	if stale.ServerUpdatedAt != baseUpdatedAt {
		t.Fatalf("stale 回执 serverUpdatedAt = %q，库中 = %q", stale.ServerUpdatedAt, baseUpdatedAt)
	}
	cur, ok := doccServerCurrent(t, stale)
	if !ok {
		t.Fatal("删除 stale 必须回带 serverCurrent（客户端据此展示败方内容）")
	}
	if cur["updatedAt"] != baseUpdatedAt {
		t.Fatalf("serverCurrent.updatedAt = %v，库中 = %q", cur["updatedAt"], baseUpdatedAt)
	}
	if cur["deletedAt"] != "" {
		t.Fatalf("stale（未删除）serverCurrent.deletedAt 必须为空，实际 = %v", cur["deletedAt"])
	}
	afterStale, deletedAfterStale := doccTaskRow(t, taskID)
	if deletedAfterStale {
		t.Fatal("base 不匹配不得产生新墓碑（deleted_at 必须仍为空）")
	}
	if afterStale != baseUpdatedAt {
		t.Fatalf("base 不匹配不得推进 updated_at：before=%q after=%q", baseUpdatedAt, afterStale)
	}

	// 3. base 命中 ⇒ applied，回执时间戳 == 库中真实值（逐字）
	applied := doccDeletePush(t, ctrl, fmt.Sprintf(
		`{"deletions":[{"table":"tasks","id":"%s","baseUpdatedAt":"%s"}]}`, taskID, baseUpdatedAt))
	if applied.Outcome != types.SyncOutcomeApplied {
		t.Fatalf("outcome = %q, want applied", applied.Outcome)
	}
	tombstoneUpdatedAt, deletedNow := doccTaskRow(t, taskID)
	if !deletedNow {
		t.Fatal("base 命中后库中必须为墓碑")
	}
	if applied.ServerUpdatedAt != tombstoneUpdatedAt {
		t.Fatalf("回执 serverUpdatedAt = %q，库中 = %q（必须逐字相等）",
			applied.ServerUpdatedAt, tombstoneUpdatedAt)
	}
	if _, ok := doccServerCurrent(t, applied); ok {
		t.Fatal("applied 不回 serverCurrent（避免体积）")
	}

	// 4. 同向重复删除（base = 墓碑版本）⇒ 幂等 applied，不推进 updated_at
	repeat := doccDeletePush(t, ctrl, fmt.Sprintf(
		`{"deletions":[{"table":"tasks","id":"%s","baseUpdatedAt":"%s"}]}`, taskID, tombstoneUpdatedAt))
	if repeat.Outcome != types.SyncOutcomeApplied {
		t.Fatalf("重复删除 outcome = %q, want applied", repeat.Outcome)
	}
	afterRepeat, _ := doccTaskRow(t, taskID)
	if afterRepeat != tombstoneUpdatedAt {
		t.Fatalf("重复删除不得推进 updated_at：before=%q after=%q", tombstoneUpdatedAt, afterRepeat)
	}
	if repeat.ServerUpdatedAt != tombstoneUpdatedAt {
		t.Fatalf("重复删除回执 serverUpdatedAt = %q，库中 = %q", repeat.ServerUpdatedAt, tombstoneUpdatedAt)
	}

	// 5. 老客户端（无 base）⇒ 删除优先（R-D 回归），回执仍是库中真实值
	const oldTaskID = "97002"
	if resp := doDerivedPush(t, ctrl, doccUID,
		fmt.Sprintf(`{"tasks":[{"id":"%s","name":"T2","state":"pending","priority":"medium"}]}`, oldTaskID),
	); resp.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("新建任务2 outcome = %q, want applied", resp.Results[0].Outcome)
	}
	old := doccDeletePush(t, ctrl, fmt.Sprintf(
		`{"deletions":[{"table":"tasks","id":"%s"}]}`, oldTaskID))
	if old.Outcome != types.SyncOutcomeApplied {
		t.Fatalf("老客户端删除 outcome = %q, want applied（删除优先逐字不变）", old.Outcome)
	}
	oldTombstoneAt, oldDeleted := doccTaskRow(t, oldTaskID)
	if !oldDeleted {
		t.Fatal("老客户端无 base 的删除必须软删")
	}
	if old.ServerUpdatedAt != oldTombstoneAt {
		t.Fatalf("老客户端删除回执 serverUpdatedAt = %q，库中 = %q", old.ServerUpdatedAt, oldTombstoneAt)
	}
	if _, ok := doccServerCurrent(t, old); ok {
		t.Fatal("applied 不回 serverCurrent")
	}
}

// mustParseDocc 解析 RFC3339Milli 时间串
func mustParseDocc(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := idutil.ParseBaseUpdatedAt(&s)
	if err != nil {
		t.Fatalf("解析时间 %q: %v", s, err)
	}
	return parsed
}
