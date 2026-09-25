// DEF-SYNC-06 契约防复发：客户端推送字段集合 ⊆ 服务端 Create/Update 可表达字段集合。
//
// 桌面端 local-first 只走 POST /sync/push（tasks 复用 CreateTaskReq）；任一客户端推送字段
// 在服务端请求结构体中缺位，都会被 Go 反序列化静默丢弃且推送仍返回成功（假成功），
// 客户端清队列后下次 pull 被服务端 NULL 覆盖 ⇒ 状态回退。本契约测试锁定该字段集合。
package controllers

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	taskApp "naotodoserver/application/task"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/interfaces/types"
)

// clientTaskPushFields 桌面端 buildTaskPush 产出的任务字段集合
// （nao-todo: packages/infrastructure/src/persistence-sync/sync-service.ts）。
// 该列表是客户端推送契约的镜像，字段缺失即回归。
var clientTaskPushFields = []string{
	"id", "parentTaskId", "name", "description", "state", "priority",
	"startAt", "endAt", "projectId", "tags",
	"archivedAt", "starMarkAt", "givenUpAt",
	"remindAt", "remindRepeat", "remindTime", "remindWeekdays",
	"sortId", "createdAt", "updatedAt", "deletedAt",
}

// jsonTagSet 收集结构体（含嵌入）全部 JSON tag 名。
func jsonTagSet(typ reflect.Type) map[string]struct{} {
	set := make(map[string]struct{})
	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				walk(f.Type)
				continue
			}
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			set[name] = struct{}{}
		}
	}
	walk(typ)
	return set
}

// TestSyncPushContract_ClientTaskFieldsExpressible 客户端任务推送字段必须全部可被
// 服务端 CreateTaskReq 或 UpdateTaskReq 表达（任一缺失即会静默丢弃）。
func TestSyncPushContract_ClientTaskFieldsExpressible(t *testing.T) {
	expressible := jsonTagSet(reflect.TypeOf(types.CreateTaskReq{}))
	for name := range jsonTagSet(reflect.TypeOf(types.UpdateTaskReq{})) {
		expressible[name] = struct{}{}
	}
	var missing []string
	for _, f := range clientTaskPushFields {
		if _, ok := expressible[f]; !ok {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("客户端推送字段在服务端 Create/Update 均无法表达（将被静默丢弃）: %v", missing)
	}
}

// TestSyncPushContract_BaseUpdatedAtNotInSharedCreateReq C-44：OCC base 字段只存在于
// sync 专用条目，不得污染共享 create DTO（create REST 的 JSON 契约不变）。
func TestSyncPushContract_BaseUpdatedAtNotInSharedCreateReq(t *testing.T) {
	for _, shared := range []reflect.Type{
		reflect.TypeOf(types.CreateTaskReq{}),
		reflect.TypeOf(types.CreateTaskCheckItemReq{}),
		reflect.TypeOf(types.CreateTaskCommentReq{}),
		reflect.TypeOf(types.CreateProjectReq{}),
		reflect.TypeOf(types.CreateTagReq{}),
		reflect.TypeOf(types.CreatePomodoroReq{}),
		reflect.TypeOf(types.CreatePomodoroRecordReq{}),
	} {
		if _, ok := jsonTagSet(shared)["baseUpdatedAt"]; ok {
			t.Fatalf("共享 create DTO %s 不应含 baseUpdatedAt（C-44）", shared.Name())
		}
	}
	syncItems := map[string]reflect.Type{
		"tasks":           reflect.TypeOf(types.SyncTaskPushItem{}),
		"taskCheckItems":  reflect.TypeOf(types.SyncCheckItemPushItem{}),
		"taskComments":    reflect.TypeOf(types.SyncCommentPushItem{}),
		"projects":        reflect.TypeOf(types.SyncProjectPushItem{}),
		"tags":            reflect.TypeOf(types.SyncTagPushItem{}),
		"pomodoros":       reflect.TypeOf(types.SyncPomodoroPushItem{}),
		"pomodoroRecords": reflect.TypeOf(types.SyncPomodoroRecordPushItem{}),
	}
	for table, typ := range syncItems {
		if _, ok := jsonTagSet(typ)["baseUpdatedAt"]; !ok {
			t.Fatalf("sync 条目 %s 缺少 baseUpdatedAt", table)
		}
	}
}

// TestSyncPushContract_BaseUpdatedAtBindingEndToEnd baseUpdatedAt 经 JSON 绑定 → app DTO → VO
// 全链透传（防中途丢弃导致 OCC 静默回退 LWW）。
func TestSyncPushContract_BaseUpdatedAtBindingEndToEnd(t *testing.T) {
	const baseStr = "2026-09-24T10:00:00.123Z"
	payload := `{"tasks":[{"id":"9001","name":"任务","state":"pending","priority":"medium",` +
		`"baseUpdatedAt":"` + baseStr + `"}]}`
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("绑定 sync push: %v", err)
	}
	if got := req.Tasks[0].BaseUpdatedAt; got != baseStr {
		t.Fatalf("baseUpdatedAt 绑定 = %q, want %q", got, baseStr)
	}
	appReq := toCreateTaskReqFromSync(&req.Tasks[0])
	appReq.BaseUpdatedAt = basePtr(req.Tasks[0].BaseUpdatedAt)
	vo, err := taskApp.CreateTaskReqToValueObject(1001, appReq)
	if err != nil {
		t.Fatalf("CreateTaskReqToValueObject: %v", err)
	}
	want, _ := time.Parse(time.RFC3339Nano, baseStr)
	if !vo.BaseUpdatedAt.Equal(want) {
		t.Fatalf("VO.BaseUpdatedAt = %v, want %v", vo.BaseUpdatedAt, want)
	}
}

// TestSyncPushContract_BaseUpdatedAtMissingFallsBack base 缺失/空串 ⇒ app DTO 为 nil ⇒
// VO 零值（回退现行 LWW，向后兼容）。
func TestSyncPushContract_BaseUpdatedAtMissingFallsBack(t *testing.T) {
	for _, body := range []string{
		`{"tasks":[{"name":"t","state":"pending","priority":"medium"}]}`,
		`{"tasks":[{"name":"t","state":"pending","priority":"medium","baseUpdatedAt":""}]}`,
	} {
		var req types.SyncPushReq
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			t.Fatalf("绑定 sync push: %v", err)
		}
		appReq := toCreateTaskReq(&req.Tasks[0].CreateTaskReq)
		appReq.BaseUpdatedAt = basePtr(req.Tasks[0].BaseUpdatedAt)
		if appReq.BaseUpdatedAt != nil {
			t.Fatalf("base 缺失/空串应回退 nil, got %q", *appReq.BaseUpdatedAt)
		}
		vo, err := taskApp.CreateTaskReqToValueObject(1001, appReq)
		if err != nil {
			t.Fatalf("CreateTaskReqToValueObject: %v", err)
		}
		if !vo.BaseUpdatedAt.IsZero() {
			t.Fatalf("base 缺失/空串应得零值, got %v", vo.BaseUpdatedAt)
		}
	}
}

// TestSyncPushContract_StatusTimestampBindingEndToEnd 状态时间戳经 sync push
// JSON 绑定 → 接口层转换 → 应用层 VO 全链透传（防再次中途丢弃）。
// T319：sync 路径的可空时间改由三态承载（NullableString 遮蔽内嵌 *string），
// 本用例随之改走 toCreateTaskReqFromSync（真实写路径），语义不变。
func TestSyncPushContract_StatusTimestampBindingEndToEnd(t *testing.T) {
	payload := `{"tasks":[{"id":"9001","name":"任务","state":"pending","priority":"medium",` +
		`"archivedAt":"2026-09-21T10:00:00Z","starMarkAt":"2026-09-21T10:01:00Z","givenUpAt":"2026-09-21T10:02:00Z"}]}`
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		t.Fatalf("绑定 sync push: %v", err)
	}
	if len(req.Tasks) != 1 {
		t.Fatalf("tasks 条数 = %d, want 1", len(req.Tasks))
	}
	vo, err := taskApp.CreateTaskReqToValueObject(1001, toCreateTaskReqFromSync(&req.Tasks[0]))
	if err != nil {
		t.Fatalf("CreateTaskReqToValueObject: %v", err)
	}
	for _, c := range []struct {
		name string
		nt   *domaintypes.NullableTime
		want string
	}{
		{"archivedAt", &vo.ArchivedAt, "2026-09-21T10:00:00Z"},
		{"starMarkAt", &vo.StarMarkAt, "2026-09-21T10:01:00Z"},
		{"givenUpAt", &vo.GivenUpAt, "2026-09-21T10:02:00Z"},
	} {
		if !c.nt.Valid {
			t.Fatalf("%s 经 push 绑定后丢失（Valid=false）", c.name)
		}
		got, ok := c.nt.Value()
		if !ok || got.UTC().Format(time.RFC3339) != c.want {
			t.Fatalf("%s 经 push 绑定后错误: ok=%v got=%v want %s", c.name, ok, got, c.want)
		}
	}
}
