// T319 单测：/sync/push 可空时间三态（absent / null / "" / 值）
//
// 缺陷背景（T318 根因）：共享 CreateTaskReq 用 `*string` 承载可空时间，JSON `null` 与「字段缺省」
// 都绑成 nil ⇒ 客户端发 `archivedAt: null`（清空）被当成「不改」静默丢弃，而覆盖分支仍写其它列并
// bump updated_at、回执 applied ⇒ 客户端出队 ⇒ 下次 pull 用服务端旧值覆盖本地（P1 同步回滚）。
//
// 修复口径（本单）：sync 专用条目对六个可空时间字段改用三态承载，JSON 名不变（客户端零改动）：
//   - absent ⇒ 不写列（⛔ 不得当清空，否则回归）
//   - null   ⇒ 显式清空（写 NULL）
//   - ""     ⇒ 显式清空（维持既有约定）
//   - 值     ⇒ 设值
package controllers

import (
	"encoding/json"
	"testing"
	"time"

	taskApp "naotodoserver/application/task"
	"naotodoserver/domain/task/entities"
	taskvalueobjects "naotodoserver/domain/task/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/interfaces/types"
)

// nullableTimeField 单个可空时间字段在「JSON 名 / sync 三态 / 内嵌 *string / 应用层 VO」四层间的映射。
type nullableTimeField struct {
	jsonName string
	sync     func(types.SyncTaskPushItem) types.NullableString
	embedded func(types.SyncTaskPushItem) *string
	vo       func(*taskvalueobjects.CreateTask) domaintypes.NullableTime
}

// nullableTimeFields 同 class 六字段（与 valueobjects.nullableTimeFromCreateReq 使用面完全一致）。
var nullableTimeFields = []nullableTimeField{
	{
		jsonName: "startAt",
		sync:     func(i types.SyncTaskPushItem) types.NullableString { return i.StartAt },
		embedded: func(i types.SyncTaskPushItem) *string { return i.CreateTaskReq.StartAt },
		vo:       func(v *taskvalueobjects.CreateTask) domaintypes.NullableTime { return v.StartAt },
	},
	{
		jsonName: "endAt",
		sync:     func(i types.SyncTaskPushItem) types.NullableString { return i.EndAt },
		embedded: func(i types.SyncTaskPushItem) *string { return i.CreateTaskReq.EndAt },
		vo:       func(v *taskvalueobjects.CreateTask) domaintypes.NullableTime { return v.EndAt },
	},
	{
		jsonName: "archivedAt",
		sync:     func(i types.SyncTaskPushItem) types.NullableString { return i.ArchivedAt },
		embedded: func(i types.SyncTaskPushItem) *string { return i.CreateTaskReq.ArchivedAt },
		vo:       func(v *taskvalueobjects.CreateTask) domaintypes.NullableTime { return v.ArchivedAt },
	},
	{
		jsonName: "starMarkAt",
		sync:     func(i types.SyncTaskPushItem) types.NullableString { return i.StarMarkAt },
		embedded: func(i types.SyncTaskPushItem) *string { return i.CreateTaskReq.StarMarkAt },
		vo:       func(v *taskvalueobjects.CreateTask) domaintypes.NullableTime { return v.StarMarkAt },
	},
	{
		jsonName: "givenUpAt",
		sync:     func(i types.SyncTaskPushItem) types.NullableString { return i.GivenUpAt },
		embedded: func(i types.SyncTaskPushItem) *string { return i.CreateTaskReq.GivenUpAt },
		vo:       func(v *taskvalueobjects.CreateTask) domaintypes.NullableTime { return v.GivenUpAt },
	},
	{
		jsonName: "remindAt",
		sync:     func(i types.SyncTaskPushItem) types.NullableString { return i.RemindAt },
		embedded: func(i types.SyncTaskPushItem) *string { return i.CreateTaskReq.RemindAt },
		vo:       func(v *taskvalueobjects.CreateTask) domaintypes.NullableTime { return v.RemindAt },
	},
}

// nullableTimeCase 三态用例：JSON 片段 + 期望的 sync 三态与应用层 NullableTime。
type nullableTimeCase struct {
	name      string
	jsonValue string // 该字段的 JSON 值片段（absent 用例不使用）
	present   bool
	null      bool
	value     string
	voValid   bool
	voNull    bool
}

var nullableTimeCases = []nullableTimeCase{
	{name: "absent", present: false, null: false, value: "", voValid: false, voNull: true},
	{name: "null", jsonValue: "null", present: true, null: true, value: "", voValid: true, voNull: true},
	{name: "empty", jsonValue: `""`, present: true, null: false, value: "", voValid: true, voNull: true},
	{
		name:      "value",
		jsonValue: `"2026-01-01T00:00:00Z"`,
		present:   true,
		null:      false,
		value:     "2026-01-01T00:00:00Z",
		voValid:   true,
		voNull:    false,
	},
}

// syncTaskPayload 构造单条 sync 任务推送 JSON（可选附加一个可空时间字段）。
func syncTaskPayload(fieldName, fieldJSONValue string) string {
	body := `{"tasks":[{"id":"9001","name":"任务","state":"pending","priority":"medium",` +
		`"updatedAt":"2026-01-01T00:00:00.000Z"`
	if fieldName != "" {
		body += `,"` + fieldName + `":` + fieldJSONValue
	}
	return body + `}]}`
}

// bindSyncTaskItem 走真实 JSON 绑定（与 gin ShouldBindJSON 同一 encoding/json 规则）。
func bindSyncTaskItem(t *testing.T, fieldName, fieldJSONValue string) types.SyncTaskPushItem {
	t.Helper()
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(syncTaskPayload(fieldName, fieldJSONValue)), &req); err != nil {
		t.Fatalf("绑定 sync push: %v", err)
	}
	if len(req.Tasks) != 1 {
		t.Fatalf("tasks 条数 = %d, want 1", len(req.Tasks))
	}
	return req.Tasks[0]
}

// TestSyncPushNullableTimeTriStateMatrix 六字段 × 四态：JSON 绑定正确落到三态承载，
// 且内嵌 *string 同名位保持 nil（证明浅层遮蔽生效、不会双重绑定）。
func TestSyncPushNullableTimeTriStateMatrix(t *testing.T) {
	for _, field := range nullableTimeFields {
		for _, tc := range nullableTimeCases {
			t.Run(field.jsonName+"/"+tc.name, func(t *testing.T) {
				name, jsonValue := "", ""
				if tc.present {
					name, jsonValue = field.jsonName, tc.jsonValue
				}
				item := bindSyncTaskItem(t, name, jsonValue)

				got := field.sync(item)
				if got.Present != tc.present || got.Null != tc.null || got.Value != tc.value {
					t.Fatalf("三态 = {Present:%v Null:%v Value:%q}, want {Present:%v Null:%v Value:%q}",
						got.Present, got.Null, got.Value, tc.present, tc.null, tc.value)
				}
				if embedded := field.embedded(item); embedded != nil {
					t.Fatalf("内嵌 *string 被 JSON 绑定写入 %q（遮蔽失效 ⇒ 双重语义）", *embedded)
				}
			})
		}
	}
}

// TestSyncPushNullableTimeReachesValueObject 三态 → 应用层 VO：
// absent 仍是「缺省不写列」（Valid=false）；null / "" 是「显式清空」（Valid=true,IsNull=true）；
// 值解析为指定时间。回归点：⛔ 不得把 absent 也当清空。
func TestSyncPushNullableTimeReachesValueObject(t *testing.T) {
	for _, field := range nullableTimeFields {
		for _, tc := range nullableTimeCases {
			t.Run(field.jsonName+"/"+tc.name, func(t *testing.T) {
				name, jsonValue := "", ""
				if tc.present {
					name, jsonValue = field.jsonName, tc.jsonValue
				}
				item := bindSyncTaskItem(t, name, jsonValue)
				appReq := toCreateTaskReqFromSync(&item)
				vo, err := taskApp.CreateTaskReqToValueObject(1001, appReq)
				if err != nil {
					t.Fatalf("CreateTaskReqToValueObject: %v", err)
				}
				got := field.vo(vo)
				if got.Valid != tc.voValid || got.IsNull != tc.voNull {
					t.Fatalf("VO = {Valid:%v IsNull:%v}, want {Valid:%v IsNull:%v}",
						got.Valid, got.IsNull, tc.voValid, tc.voNull)
				}
				if tc.voValid && !tc.voNull {
					want, _ := time.Parse(time.RFC3339, tc.value)
					if !got.Time.Equal(want) {
						t.Fatalf("VO.Time = %v, want %v", got.Time, want)
					}
				}
			})
		}
	}
}

// TestSyncPushClientRealPayloadClearsNullableTimes T318 实测载荷：桌面端 v1.12.0 的
// buildTaskPush 对已取消归档的任务发 `"archivedAt":null`（键存在）——修复后必须映射为
// 「显式清空」，即应用层 VO 的 ArchivedAt.Valid=true && IsNull=true（⇒ 写 NULL）。
// 用例同时覆盖客户端同一载荷里的 starMarkAt/givenUpAt/remindAt null。
func TestSyncPushClientRealPayloadClearsNullableTimes(t *testing.T) {
	payload := `{"tasks":[{"id":"9001","name":"任务","description":"","state":"pending",` +
		`"priority":"medium","startAt":null,"endAt":null,"projectId":"","tags":[],` +
		`"archivedAt":null,"starMarkAt":null,"givenUpAt":null,"remindAt":null,` +
		`"remindRepeat":"none","remindTime":null,"remindWeekdays":[],` +
		`"createdAt":"2025-12-01T00:00:00.000Z","updatedAt":"2026-01-01T00:00:00.000Z",` +
		`"deletedAt":null,"baseUpdatedAt":"2025-12-31T23:59:59.000Z"}]}`
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("绑定客户端真实载荷: %v", err)
	}
	appReq := toCreateTaskReqFromSync(&req.Tasks[0])
	vo, err := taskApp.CreateTaskReqToValueObject(1001, appReq)
	if err != nil {
		t.Fatalf("CreateTaskReqToValueObject: %v", err)
	}
	for _, field := range nullableTimeFields {
		got := field.vo(vo)
		if !got.Valid || !got.IsNull {
			t.Fatalf("%s: null 应映射为显式清空 {Valid:true IsNull:true}, got {Valid:%v IsNull:%v}",
				field.jsonName, got.Valid, got.IsNull)
		}
	}
}

// TestSyncPushAbsentNullableTimeNeverClears 显式回归锁：载荷**不含**可空时间键时，
// 三态为 absent ⇒ 应用层入参为 nil（缺省不写列），⛔ 不得变成清空。
func TestSyncPushAbsentNullableTimeNeverClears(t *testing.T) {
	item := bindSyncTaskItem(t, "", "")
	appReq := toCreateTaskReqFromSync(&item)
	for _, ptr := range []*string{
		appReq.StartAt, appReq.EndAt, appReq.ArchivedAt,
		appReq.StarMarkAt, appReq.GivenUpAt, appReq.RemindAt,
	} {
		if ptr != nil {
			t.Fatalf("absent 字段映射出非 nil *string = %q（会被当作清空 ⇒ 回归）", *ptr)
		}
	}
}

// TestCreateRestNullableTimeUnchanged REST 契约回归：共享 CreateTaskReq/UpdateTaskReq 仍是
// `*string` 三态（nil=缺省不写、""=清空），JSON null 与字段缺省同绑 nil ⇒ REST 行为与本单前一致。
func TestCreateRestNullableTimeUnchanged(t *testing.T) {
	const payload = `{"name":"任务","state":"pending","priority":"medium","archivedAt":null}`
	var restReq types.CreateTaskReq
	if err := json.Unmarshal([]byte(payload), &restReq); err != nil {
		t.Fatalf("绑定 create REST 载荷: %v", err)
	}
	if restReq.ArchivedAt != nil {
		t.Fatalf("REST CreateTaskReq.ArchivedAt 应为 nil（null=缺省），got %q", *restReq.ArchivedAt)
	}
	vo, err := taskApp.CreateTaskReqToValueObject(1001, toCreateTaskReq(&restReq))
	if err != nil {
		t.Fatalf("CreateTaskReqToValueObject: %v", err)
	}
	if vo.ArchivedAt.Valid {
		t.Fatalf("REST null 不得触发写列（Valid 应为 false），got %+v", vo.ArchivedAt)
	}
}

// TestPullResNullableTimeUnchanged pull 出参回归：未归档任务的 archivedAt 仍输出 ""
// （非 null，客户端 ” → null）；已归档任务输出 RFC3339 秒级时间串。
func TestPullResNullableTimeUnchanged(t *testing.T) {
	unarchived := taskApp.TaskEntityToGetRes(&entities.Task{
		ArchivedAt: domaintypes.NewNullableTimeNull(),
	})
	if unarchived.ArchivedAt != "" {
		t.Fatalf("未归档 archivedAt = %q, want \"\"", unarchived.ArchivedAt)
	}
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	archived := taskApp.TaskEntityToGetRes(&entities.Task{
		ArchivedAt: domaintypes.NewNullableTimeByTime(at),
	})
	if archived.ArchivedAt != at.Format(time.RFC3339) {
		t.Fatalf("已归档 archivedAt = %q, want %q", archived.ArchivedAt, at.Format(time.RFC3339))
	}
}
