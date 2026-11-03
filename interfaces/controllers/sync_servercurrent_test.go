// T642 / A4 契约测试：/sync/push 的 `serverCurrent`（additive）与 deletions 的 `baseUpdatedAt`（OCC）。
//
// 覆盖：
//   - serverCurrent 出现条件（stale / noop 有、applied 无）
//   - serverCurrent 形状与 /sync/pull 同表条目一致（同一套 converter，字段名逐一断言）
//   - baseUpdatedAt 缺省时不出现在 JSON（omitempty）、非法值的逐条失败语义
//   - 删除回执的 serverUpdatedAt = 库中真实值（不再复用循环前的 now，R-20）
//   - 老客户端（无 base）⇒ 零值 base 透传（= 现行「删除优先」，R-D）
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// scFakeTaskApp /sync/push 契约测试桩：只实现被调用到的方法（其余由内嵌接口占位，被调用即 panic）
type scFakeTaskApp struct {
	taskApp.TaskApp
	upsert domaintypes.UpsertResult
	// current 「库中当前行」（GetTaskById / DeleteTask 回传；nil = 读取失败）
	current *taskDto.GetTaskRes
	// decision 删除裁决（DeleteTask 回传）
	decision domaintypes.DeleteDecision
	// 记录删除调用实参
	deleteCalls int
	deleteID    string
	deleteBase  time.Time
}

func (f *scFakeTaskApp) CreateTask(
	_ context.Context, _ int64, _ *taskDto.CreateTaskReq,
) (*taskDto.GetTaskRes, domaintypes.UpsertResult, error) {
	return &taskDto.GetTaskRes{Id: "9001", UpdatedAt: "2026-11-03T10:00:00.123Z"}, f.upsert, nil
}

func (f *scFakeTaskApp) GetTaskById(
	_ context.Context, _ int64, id string, _ bool,
) (*taskDto.GetTaskRes, error) {
	if f.current == nil {
		return nil, errors.New("record not found")
	}
	row := *f.current
	row.Id = id
	return &row, nil
}

func (f *scFakeTaskApp) DeleteTask(
	_ context.Context, _ int64, id string, baseUpdatedAt time.Time,
) (*taskDto.GetTaskRes, domaintypes.DeleteDecision, error) {
	f.deleteCalls++
	f.deleteID = id
	f.deleteBase = baseUpdatedAt
	return f.current, f.decision, nil
}

// pushSyncRaw 调用真实 Push 控制器，同时返回解析后的回执与原始 JSON（断言字段缺省用）
func pushSyncRaw(t *testing.T, c *SyncController, body string) (types.SyncPushRes, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/sync/push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(iCtx.SetUserId(req.Context(), 1001))
	gc.Request = req
	c.Push(gc)
	if w.Code != http.StatusOK {
		t.Fatalf("Push HTTP 状态 = %d, body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data types.SyncPushRes `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	return resp.Data, w.Body.String()
}

// firstResultRaw 取原始 JSON 中 results[0] 的键值（断言「字段是否出现」用）
func firstResultRaw(t *testing.T, raw string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("解析原始 JSON 失败: %v, raw=%s", err, raw)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("data 缺失: %s", raw)
	}
	results, ok := data["results"].([]any)
	if !ok || len(results) == 0 {
		t.Fatalf("results 缺失: %s", raw)
	}
	result, ok := results[0].(map[string]any)
	if !ok {
		t.Fatalf("results[0] 形态异常: %s", raw)
	}
	return result
}

// jsonKeys 取结构体序列化后的键集合（形状比对用）
func jsonKeys(t *testing.T, v any) map[string]struct{} {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal 失败: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal 失败: %v", err)
	}
	keys := make(map[string]struct{}, len(m))
	for k := range m {
		keys[k] = struct{}{}
	}
	return keys
}

const scTaskPushBody = `{"tasks":[{"id":"9001","name":"t","state":"pending","priority":"medium",` +
	`"createdAt":"2026-11-03T09:00:00Z","updatedAt":"2026-11-03T10:00:00Z"}]}`

// TestSyncPushServerCurrent_PresenceByOutcome serverCurrent 仅对 stale / noop 返回（applied 不回，避免体积）。
func TestSyncPushServerCurrent_PresenceByOutcome(t *testing.T) {
	cases := []struct {
		name   string
		upsert domaintypes.UpsertResult
		want   bool
	}{
		{"applied（已写入）⇒ 不回当前行", domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite}, false},
		{"noop（LWW 拒）⇒ 回当前行", domaintypes.UpsertResult{Outcome: domaintypes.UpsertNoop}, true},
		{"stale（OCC 不匹配）⇒ 回当前行", domaintypes.UpsertResult{Outcome: domaintypes.UpsertStale}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := &scFakeTaskApp{
				upsert:  tc.upsert,
				current: &taskDto.GetTaskRes{Id: "9001", UpdatedAt: "2026-11-03T10:00:00.123Z", Name: "云端名称"},
			}
			c := NewSyncController(app, nil, nil, nil, nil, nil)
			got, raw := pushSyncRaw(t, c, scTaskPushBody)

			result := firstResultRaw(t, raw)
			_, present := result["serverCurrent"]
			if present != tc.want {
				t.Fatalf("serverCurrent 出现 = %v, want %v (raw=%s)", present, tc.want, raw)
			}
			if tc.want != (got.Results[0].ServerCurrent != nil) {
				t.Fatalf("解析后 ServerCurrent = %#v, want 非空 = %v", got.Results[0].ServerCurrent, tc.want)
			}
			if !tc.want {
				return
			}
			cur, ok := result["serverCurrent"].(map[string]any)
			if !ok {
				t.Fatalf("serverCurrent 不是对象: %s", raw)
			}
			if cur["name"] != "云端名称" {
				t.Fatalf("serverCurrent.name = %v, want 云端名称", cur["name"])
			}
			if cur["updatedAt"] != "2026-11-03T10:00:00.123Z" {
				t.Fatalf("serverCurrent.updatedAt = %v", cur["updatedAt"])
			}
		})
	}
}

// TestSyncPushServerCurrent_ShapeMatchesPullItem 形状必须与 /sync/pull 同表条目一致（同一套 converter）。
func TestSyncPushServerCurrent_ShapeMatchesPullItem(t *testing.T) {
	app := &scFakeTaskApp{
		upsert: domaintypes.UpsertResult{Outcome: domaintypes.UpsertStale},
		current: &taskDto.GetTaskRes{
			Id: "9001", CreatedAt: "2026-11-03T09:00:00Z", UpdatedAt: "2026-11-03T10:00:00.123Z",
			DeletedAt: "", Name: "n", Description: "d", State: "pending", Priority: "medium",
			Tags: []string{}, SortId: 256,
		},
	}
	c := NewSyncController(app, nil, nil, nil, nil, nil)
	_, raw := pushSyncRaw(t, c, scTaskPushBody)

	cur, ok := firstResultRaw(t, raw)["serverCurrent"].(map[string]any)
	if !ok {
		t.Fatalf("serverCurrent 缺失: %s", raw)
	}
	// pull 条目的形状来源（同一套 converter）：toGetTaskRes
	pullKeys := jsonKeys(t, toGetTaskRes(&taskDto.GetTaskRes{Id: "9001"}))
	for k := range pullKeys {
		if _, ok := cur[k]; !ok {
			t.Fatalf("serverCurrent 缺少 pull 条目字段 %q（形状必须与 /sync/pull 同源）", k)
		}
	}
	for _, k := range []string{"id", "updatedAt", "deletedAt"} {
		if _, ok := cur[k]; !ok {
			t.Fatalf("serverCurrent 必须含 %q（客户端据此判断墓碑与 base）", k)
		}
	}
}

// TestSyncDeletion_BaseUpdatedAtOmitempty 缺省时 JSON 不出现该键（老客户端无感）。
func TestSyncDeletion_BaseUpdatedAtOmitempty(t *testing.T) {
	raw, err := json.Marshal(types.SyncDeletion{Table: "tasks", Id: "1"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "baseUpdatedAt") {
		t.Fatalf("缺省 base 不应出现在 JSON，实际 = %s", raw)
	}
	base := "2026-11-03T10:00:00.123Z"
	raw, err = json.Marshal(types.SyncDeletion{Table: "tasks", Id: "1", BaseUpdatedAt: &base})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"baseUpdatedAt":"2026-11-03T10:00:00.123Z"`) {
		t.Fatalf("显式 base 必须出现，实际 = %s", raw)
	}
}

// TestSyncPushDeletion_RealTimestampAndOCC 删除回执的 serverUpdatedAt 必须是库中真实值，
// 并按 base（缺省/命中/不匹配/墓碑幂等）给出 outcome + 必要的 serverCurrent。
func TestSyncPushDeletion_RealTimestampAndOCC(t *testing.T) {
	const dbUpdatedAt = "2026-11-03T10:00:00.456Z"
	tombstone := &taskDto.GetTaskRes{
		Id: "9001", UpdatedAt: dbUpdatedAt, DeletedAt: dbUpdatedAt, Name: "已删除",
	}
	alive := &taskDto.GetTaskRes{Id: "9001", UpdatedAt: dbUpdatedAt, DeletedAt: "", Name: "还活着"}

	cases := []struct {
		name         string
		body         string
		app          *scFakeTaskApp
		wantOutcome  string
		wantBaseZero bool
		wantCurrent  bool
	}{
		{
			name:         "老客户端（无 base）⇒ 删除优先，applied，无 serverCurrent",
			body:         `{"deletions":[{"table":"tasks","id":"9001"}]}`,
			app:          &scFakeTaskApp{current: tombstone, decision: domaintypes.DeleteApply},
			wantOutcome:  types.SyncOutcomeApplied,
			wantBaseZero: true,
		},
		{
			name:        "base 命中 ⇒ applied",
			body:        `{"deletions":[{"table":"tasks","id":"9001","baseUpdatedAt":"` + dbUpdatedAt + `"}]}`,
			app:         &scFakeTaskApp{current: tombstone, decision: domaintypes.DeleteApply},
			wantOutcome: types.SyncOutcomeApplied,
		},
		{
			name:        "base 不匹配 ⇒ stale + serverCurrent（服务端当前行）",
			body:        `{"deletions":[{"table":"tasks","id":"9001","baseUpdatedAt":"2026-11-03T09:00:00Z"}]}`,
			app:         &scFakeTaskApp{current: alive, decision: domaintypes.DeleteStale},
			wantOutcome: types.SyncOutcomeStale,
			wantCurrent: true,
		},
		{
			name:        "已是墓碑（同向重复删除）⇒ 幂等 applied，无 serverCurrent",
			body:        `{"deletions":[{"table":"tasks","id":"9001","baseUpdatedAt":"` + dbUpdatedAt + `"}]}`,
			app:         &scFakeTaskApp{current: tombstone, decision: domaintypes.DeleteIdempotent},
			wantOutcome: types.SyncOutcomeApplied,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewSyncController(tc.app, nil, nil, nil, nil, nil)
			got, raw := pushSyncRaw(t, c, tc.body)
			if len(got.Results) != 1 {
				t.Fatalf("results 条数 = %d, want 1", len(got.Results))
			}
			r := got.Results[0]
			if r.Outcome != tc.wantOutcome {
				t.Fatalf("outcome = %q, want %q", r.Outcome, tc.wantOutcome)
			}
			// R-20：回执时间戳 = 库中真实 updated_at（逐字），不是循环前的 now
			if r.ServerUpdatedAt != dbUpdatedAt {
				t.Fatalf("serverUpdatedAt = %q, want 库中真实值 %q", r.ServerUpdatedAt, dbUpdatedAt)
			}
			if tc.wantBaseZero {
				if !tc.app.deleteBase.IsZero() {
					t.Fatalf("无 base 时应透传零值（回退删除优先），实际 = %v", tc.app.deleteBase)
				}
			} else if tc.app.deleteBase.IsZero() {
				t.Fatalf("携带 base 时应解析后透传，实际为零值（raw=%s）", raw)
			}
			_, hasCurrent := firstResultRaw(t, raw)["serverCurrent"]
			if hasCurrent != tc.wantCurrent {
				t.Fatalf("serverCurrent 出现 = %v, want %v (raw=%s)", hasCurrent, tc.wantCurrent, raw)
			}
		})
	}
}

// TestSyncPushDeletion_InvalidBaseIsPerItemError 非法 base 只影响该条（逐条 error=invalid），不整批 400。
func TestSyncPushDeletion_InvalidBaseIsPerItemError(t *testing.T) {
	app := &scFakeTaskApp{current: &taskDto.GetTaskRes{Id: "9001"}, decision: domaintypes.DeleteApply}
	c := NewSyncController(app, nil, nil, nil, nil, nil)
	got, _ := pushSyncRaw(t, c, `{"deletions":[{"table":"tasks","id":"9001","baseUpdatedAt":"not-a-time"}]}`)
	if len(got.Results) != 1 {
		t.Fatalf("results 条数 = %d, want 1", len(got.Results))
	}
	r := got.Results[0]
	if r.Outcome != types.SyncOutcomeError || r.Error != "invalid" {
		t.Fatalf("outcome/error = %q/%q, want error/invalid", r.Outcome, r.Error)
	}
	if app.deleteCalls != 0 {
		t.Fatalf("非法 base 不得触达删除逻辑，deleteCalls = %d", app.deleteCalls)
	}
}

// TestSyncPushDeletion_PomodoroRecordStillSkipped 只追加资源（pomodoroRecords）删除语义不变。
func TestSyncPushDeletion_PomodoroRecordStillSkipped(t *testing.T) {
	c := NewSyncController(&scFakeTaskApp{}, nil, nil, nil, nil, nil)
	got, _ := pushSyncRaw(t, c, `{"deletions":[{"table":"pomodoroRecords","id":"1"}]}`)
	if len(got.Results) != 1 {
		t.Fatalf("results 条数 = %d, want 1", len(got.Results))
	}
	r := got.Results[0]
	if !r.Skipped || r.Outcome != types.SyncOutcomeSkipped {
		t.Fatalf("skipped/outcome = %v/%q, want true/skipped", r.Skipped, r.Outcome)
	}
}
