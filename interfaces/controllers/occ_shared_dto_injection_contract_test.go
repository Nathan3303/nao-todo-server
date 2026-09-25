// T163b 契约守护：共享 create DTO 不得成为 OCC base 的 JSON 注入面。
//
// 背景：OCC 的 `baseUpdatedAt` 仅属于 sync 通道，只应由 `/sync/push` 控制器在 Go 侧
// 显式赋值。若它被加在共享 create DTO 上却不带 JSON tag，`encoding/json` 会按字段名
// **大小写不敏感**匹配把它从请求体注入 ⇒ `POST /tasks` 等 create REST 端点可触发 OCC/stale。
//
// 本文件的两类断言都具备**判别力**（修复前会打红）：
//   - TestSharedCreateDTO_BaseUpdatedAtNotJSONBindable：直接 json.Unmarshal 注入载荷
//   - TestCreateTaskREST_BaseUpdatedAtIgnored：真实 create 控制器 + 携带字段的请求体
package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	pomodoroDto "naotodoserver/application/pomodoro/dto"
	projectDto "naotodoserver/application/project/dto"
	tagDto "naotodoserver/application/tag/dto"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"

	"github.com/gin-gonic/gin"
)

// sharedCreateDTOInstances 7 个共享 create DTO 的零值实例（应用层内部 DTO，无 JSON tag 契约）
func sharedCreateDTOInstances() map[string]any {
	return map[string]any{
		"task.CreateTaskReq":               &taskDto.CreateTaskReq{},
		"task.CreateTaskCheckItemReq":      &taskDto.CreateTaskCheckItemReq{},
		"task.CreateTaskCommentReq":        &taskDto.CreateTaskCommentReq{},
		"tag.CreateTagReq":                 &tagDto.CreateTagReq{},
		"pomodoro.CreatePomodoroReq":       &pomodoroDto.CreatePomodoroReq{},
		"pomodoro.CreatePomodoroRecordReq": &pomodoroDto.CreatePomodoroRecordReq{},
		"project.CreateProjectReq":         &projectDto.CreateProjectReq{},
	}
}

// baseUpdatedAtField 取实例的 BaseUpdatedAt 指针字段（缺失即 fail）
func baseUpdatedAtField(t *testing.T, name string, ptr any) reflect.Value {
	t.Helper()
	field := reflect.ValueOf(ptr).Elem().FieldByName("BaseUpdatedAt")
	if !field.IsValid() {
		t.Fatalf("%s: 缺少 BaseUpdatedAt 字段", name)
	}
	if field.Kind() != reflect.Pointer {
		t.Fatalf("%s: BaseUpdatedAt 应为指针, got %s", name, field.Kind())
	}
	return field
}

// TestSharedCreateDTO_BaseUpdatedAtNotJSONBindable 判别力断言：
// 共享 create DTO 经 JSON 反序列化（含大小写变体）后，BaseUpdatedAt 必须仍为 nil。
// 修复前（字段无 tag）本用例会因被注入而打红。
func TestSharedCreateDTO_BaseUpdatedAtNotJSONBindable(t *testing.T) {
	payload := []byte(`{"baseUpdatedAt":"2026-09-24T10:00:00.123Z",` +
		`"BASEUPDATEDAT":"2026-09-24T11:00:00.456Z",` +
		`"BaseUpdatedAt":"2026-09-24T12:00:00.789Z"}`)
	for name, ptr := range sharedCreateDTOInstances() {
		if err := json.Unmarshal(payload, ptr); err != nil {
			t.Fatalf("%s: json.Unmarshal: %v", name, err)
		}
		if got := baseUpdatedAtField(t, name, ptr).Interface().(*string); got != nil {
			t.Fatalf("%s: baseUpdatedAt 经 JSON 注入成功（%q）—— 共享 create DTO 存在 OCC 注入面", name, *got)
		}
	}
}

// TestSharedCreateDTO_BaseUpdatedAtJSONTagClosed 结构性断言：字段须显式关闭 JSON 绑定。
func TestSharedCreateDTO_BaseUpdatedAtJSONTagClosed(t *testing.T) {
	for name, ptr := range sharedCreateDTOInstances() {
		field, ok := reflect.TypeOf(ptr).Elem().FieldByName("BaseUpdatedAt")
		if !ok {
			t.Fatalf("%s: 缺少 BaseUpdatedAt 字段", name)
		}
		if got := field.Tag.Get("json"); got != "-" {
			t.Fatalf("%s: BaseUpdatedAt json tag = %q, want \"-\"（防 JSON 注入面）", name, got)
		}
	}
}

// captureTaskApp 捕获 create REST 传给应用层的任务 DTO
type captureTaskApp struct {
	taskApp.TaskApp
	got *taskDto.CreateTaskReq
}

func (f *captureTaskApp) CreateTask(
	_ context.Context, _ int64, req *taskDto.CreateTaskReq,
) (*taskDto.GetTaskRes, domaintypes.UpsertResult, error) {
	f.got = req
	return &taskDto.GetTaskRes{Id: "9001", UpdatedAt: "2026-09-24T00:00:00Z"},
		domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite, Created: true}, nil
}

// postCreateTask 以已登录用户身份调用真实 POST /tasks 控制器，返回捕获的 app DTO
func postCreateTask(t *testing.T, body string) *taskDto.CreateTaskReq {
	t.Helper()
	app := &captureTaskApp{}
	c := NewTaskController(app)

	gc := newGinContext(t, http.MethodPost, "/tasks", body)
	c.CreateTask(gc)
	if app.got == nil {
		t.Fatal("create 控制器未调用应用层 CreateTask")
	}
	return app.got
}

// TestCreateTaskREST_BaseUpdatedAtIgnored 行为断言：create REST 请求体携带 baseUpdatedAt
// ⇒ 不影响 app DTO（BaseUpdatedAt 保持 nil），即不产生 OCC/stale，行为与不携带逐字一致。
func TestCreateTaskREST_BaseUpdatedAtIgnored(t *testing.T) {
	withBase := postCreateTask(t, `{"name":"t","state":"pending","priority":"medium",`+
		`"baseUpdatedAt":"2026-09-24T10:00:00.123Z"}`)
	without := postCreateTask(t, `{"name":"t","state":"pending","priority":"medium"}`)

	if withBase.BaseUpdatedAt != nil {
		t.Fatalf("POST /tasks 注入 baseUpdatedAt 成功（%q）—— create REST 不应触发 OCC", *withBase.BaseUpdatedAt)
	}
	if without.BaseUpdatedAt != nil {
		t.Fatalf("对照请求 BaseUpdatedAt 应为 nil, got %q", *without.BaseUpdatedAt)
	}
	// 行为一致：携带与不携带 baseUpdatedAt 得到同一 DTO 值
	if !reflect.DeepEqual(withBase, without) {
		t.Fatalf("携带/不携带 baseUpdatedAt 的 app DTO 不一致:\n with=%+v\nwithout=%+v", withBase, without)
	}
}

// newGinContext 构造已登录用户的 gin 测试上下文（含 JSON body）
func newGinContext(t *testing.T, method, target, body string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(iCtx.SetUserId(req.Context(), 1001))
	gc.Request = req
	return gc
}
