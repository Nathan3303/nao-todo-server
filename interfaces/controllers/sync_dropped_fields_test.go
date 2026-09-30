// T466 单测：/sync/push「字段被静默丢弃」护栏（DEF-44，additive DROPPED fields）。
//
// 背景：服务端只承载 sync 条目 DTO 声明的字段，载荷中其余键被 encoding/json 静默丢弃，
// 却仍回执 applied ⇒ 客户端出队并误以为已同步。
//
// 口径（本单）：对每条推送记录比对「原始 JSON 键集 − sync 条目 DTO 可承载键集（含嵌入）」，
// 差集按字典序写入 `SyncResult.DroppedFields`；**outcome 语义逐字不变**（additive）。
package controllers

import (
	"encoding/json"
	"reflect"
	"testing"

	"naotodoserver/interfaces/types"
)

// TestSyncDroppedFields_UncarriedKeysFlagged 载荷含未承载字段 ⇒ 逐条上报（字典序）。
// 用例字段：projects 的 `icon`（未承载）与 `sortId`（未承载，DEF-47 族）。
func TestSyncDroppedFields_UncarriedKeysFlagged(t *testing.T) {
	d := newSyncDroppedFields([]byte(
		`{"projects":[{"name":"清单","description":"","icon":"star","sortId":3}]}`))
	got := d.at("projects", 0)
	want := []string{"icon", "sortId"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("droppedFields = %v, want %v", got, want)
	}
}

// TestSyncDroppedFields_CarriedPayloadSilent 全部键均被承载 ⇒ 不产生告警（避免噪声）。
func TestSyncDroppedFields_CarriedPayloadSilent(t *testing.T) {
	d := newSyncDroppedFields([]byte(
		`{"projects":[{"id":"1","name":"清单","description":"","deletedAt":null,` +
			`"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z",` +
			`"archivedAt":null,"deactivedAt":null,"baseUpdatedAt":"2026-01-01T00:00:00Z"}]}`))
	if got := d.at("projects", 0); got != nil {
		t.Fatalf("已承载字段不应告警, got %v", got)
	}
}

// TestSyncDroppedFields_NullValuedUncarriedStillFlagged 键存在即算「载荷含该字段」，
// 与值是否为 null 无关（null 同样是客户端表达写意图）。
func TestSyncDroppedFields_NullValuedUncarriedStillFlagged(t *testing.T) {
	d := newSyncDroppedFields([]byte(`{"projects":[{"name":"清单","icon":null}]}`))
	if got := d.at("projects", 0); !reflect.DeepEqual(got, []string{"icon"}) {
		t.Fatalf("null 值未承载字段应告警, got %v", got)
	}
}

// TestSyncDroppedFields_IndexAlignsPerTable 同表多条按出现顺序对齐；删除墓碑不占用 create 下标。
func TestSyncDroppedFields_IndexAlignsPerTable(t *testing.T) {
	d := newSyncDroppedFields([]byte(
		`{"projects":[{"name":"a","icon":"x"},{"name":"b"}],` +
			`"deletions":[{"table":"projects","id":"1"}]}`))
	if got := d.at("projects", 0); !reflect.DeepEqual(got, []string{"icon"}) {
		t.Fatalf("第 0 条 droppedFields = %v, want [icon]", got)
	}
	if got := d.at("projects", 1); got != nil {
		t.Fatalf("第 1 条无丢弃字段, got %v", got)
	}
	if got := d.at("deletions", 0); got != nil {
		t.Fatalf("删除墓碑不参与护栏, got %v", got)
	}
}

// TestSyncDroppedFields_AttachByTableOrder attach 按表内出现顺序回填，跨表互不干扰。
func TestSyncDroppedFields_AttachByTableOrder(t *testing.T) {
	d := newSyncDroppedFields([]byte(
		`{"projects":[{"name":"a","icon":"x"},{"name":"b","icon":"y"}],` +
			`"tags":[{"name":"g","color":"#fff","sortId":1}]}`))
	results := []types.SyncResult{
		{Table: "projects", Id: "1"},
		{Table: "projects", Id: "2"},
		{Table: "tags", Id: "3"},
	}
	d.attach(results)
	if !reflect.DeepEqual(results[0].DroppedFields, []string{"icon"}) {
		t.Fatalf("projects[0] = %v", results[0].DroppedFields)
	}
	if !reflect.DeepEqual(results[1].DroppedFields, []string{"icon"}) {
		t.Fatalf("projects[1] = %v", results[1].DroppedFields)
	}
	if !reflect.DeepEqual(results[2].DroppedFields, []string{"sortId"}) {
		t.Fatalf("tags[0] = %v", results[2].DroppedFields)
	}
}

// TestSyncDroppedFields_MalformedBodyDegrades 解析失败静默降级（不阻断主流程）。
func TestSyncDroppedFields_MalformedBodyDegrades(t *testing.T) {
	d := newSyncDroppedFields([]byte(`not-json`))
	if got := d.at("projects", 0); got != nil {
		t.Fatalf("非法载荷应降级为空, got %v", got)
	}
}

// TestSyncPush_DroppedFieldsEndToEnd 全链：控制器 Push 回执携带 droppedFields，
// 且 **outcome 仍为 applied**（additive：不改既有判定语义）。
func TestSyncPush_DroppedFieldsEndToEnd(t *testing.T) {
	app := &fakeProjectApp{}
	c := NewSyncController(nil, nil, nil, app, nil, nil)
	got := pushSync(t, c, `{"projects":[{"name":"清单","icon":"star"}]}`)
	if len(got.Results) != 1 {
		t.Fatalf("results 条数 = %d, want 1", len(got.Results))
	}
	r := got.Results[0]
	if r.Outcome != types.SyncOutcomeApplied {
		t.Fatalf("outcome = %q, want %q（additive 不改既有语义）", r.Outcome, types.SyncOutcomeApplied)
	}
	if !reflect.DeepEqual(r.DroppedFields, []string{"icon"}) {
		t.Fatalf("droppedFields = %v, want [icon]", r.DroppedFields)
	}
}

// TestSyncPush_DroppedFieldsOmittedWhenEmpty 无丢弃 ⇒ JSON 省略该键（旧客户端对新增字段无感）。
func TestSyncPush_DroppedFieldsOmittedWhenEmpty(t *testing.T) {
	app := &fakeProjectApp{}
	c := NewSyncController(nil, nil, nil, app, nil, nil)
	got := pushSync(t, c, `{"projects":[{"name":"清单","description":""}]}`)
	if len(got.Results) != 1 {
		t.Fatalf("results 条数 = %d, want 1", len(got.Results))
	}
	if got.Results[0].DroppedFields != nil {
		t.Fatalf("无丢弃时不应带 droppedFields, got %v", got.Results[0].DroppedFields)
	}
	encoded, err := json.Marshal(got.Results[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(encoded) != "" && containsKey(encoded, "droppedFields") {
		t.Fatalf("空 droppedFields 应从 JSON 省略, got %s", encoded)
	}
}

// containsKey 粗判 JSON 对象文本是否含指定键（仅测试辅助，避免引入额外解析）。
func containsKey(encoded []byte, key string) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
