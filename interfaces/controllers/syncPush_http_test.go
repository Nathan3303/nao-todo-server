package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// fakeSyncTaskApp 只实现 Push 用到的 CreateTask；其余方法由嵌入接口兜底（本测试不会调用）。
type fakeSyncTaskApp struct {
	taskApp.TaskApp
}

func (f *fakeSyncTaskApp) CreateTask(
	_ context.Context, _ int64, _ *taskDto.CreateTaskReq,
) (*taskDto.GetTaskRes, domaintypes.UpsertResult, error) {
	return &taskDto.GetTaskRes{
		Id:        "1",
		UpdatedAt: "2026-10-27T00:00:00.000Z",
	}, domaintypes.UpsertResult{}, nil
}

// newSyncPushEngine 用真实 SyncController 挂载 /sync/push（无 DB：app 层用 fake）。
// 同时挂上全局请求体上限中间件，覆盖 SEC-01 的「中间件 → 控制器」完整链路。
func newSyncPushEngine(fake *fakeSyncTaskApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middlewares.BodyLimit(2<<20, map[string]int64{
		"/api/users/avatar": 5 << 20,
	}))
	ctrl := controllers.NewSyncController(fake, nil, nil, nil, nil, nil)
	engine.POST("/api/sync/push", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.Push(ctx)
	})
	return engine
}

// syncTaskPushItem 构造一条最小合法 tasks 条目
func syncTaskPushItem(name string) types.SyncTaskPushItem {
	return types.SyncTaskPushItem{CreateTaskReq: types.CreateTaskReq{
		Name: name, State: "TODO", Priority: "NONE",
	}}
}

// TestSyncPushNormalPathReturnsOK 正常路径回归：≤ 上限的批量 ⇒ 200 / 90010
func TestSyncPushNormalPathReturnsOK(t *testing.T) {
	engine := newSyncPushEngine(&fakeSyncTaskApp{})
	body := types.SyncPushReq{Tasks: []types.SyncTaskPushItem{syncTaskPushItem("正常")}}

	rec := postSyncPush(t, engine, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200（body=%s）", rec.Code, rec.Body.String())
	}
	res := decodeSyncPushRes(t, rec)
	if res.Code != 90010 {
		t.Fatalf("code = %d, want 90010", res.Code)
	}
	if len(res.Data.Results) != 1 || res.Data.Results[0].Outcome != types.SyncOutcomeApplied {
		t.Fatalf("results = %+v, want 1 条 applied", res.Data.Results)
	}
}

// TestSyncPushBoundary500ReturnsOK 边界：恰好 500 条 ⇒ 仍按正常路径 200（不误拒合法请求）
func TestSyncPushBoundary500ReturnsOK(t *testing.T) {
	engine := newSyncPushEngine(&fakeSyncTaskApp{})
	tasks := make([]types.SyncTaskPushItem, 500)
	for i := range tasks {
		tasks[i] = syncTaskPushItem("边界")
	}

	rec := postSyncPush(t, engine, types.SyncPushReq{Tasks: tasks})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200（500 条应放行）", rec.Code)
	}
	if got := len(decodeSyncPushRes(t, rec).Data.Results); got != 500 {
		t.Fatalf("results = %d, want 500", got)
	}
}

// TestSyncPushOverLimitBatchReturns400 批量 501 条 ⇒ 400 + 可读错误（不逐条落库）
func TestSyncPushOverLimitBatchReturns400(t *testing.T) {
	engine := newSyncPushEngine(&fakeSyncTaskApp{})
	tasks := make([]types.SyncTaskPushItem, 501)
	for i := range tasks {
		tasks[i] = syncTaskPushItem("超限")
	}

	rec := postSyncPush(t, engine, types.SyncPushReq{Tasks: tasks})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	res := decodeSyncPushRes(t, rec)
	if res.Code != 90011 {
		t.Fatalf("code = %d, want 90011", res.Code)
	}
	for _, want := range []string{"tasks", "500", "501"} {
		if !strings.Contains(res.Error, want) {
			t.Fatalf("error 缺少 %q，实际: %q", want, res.Error)
		}
	}
}

// TestSyncPushOversizeBodyReturns413 声明 10 MB body ⇒ 413（中间件在绑定之前拦截）
func TestSyncPushOversizeBodyReturns413(t *testing.T) {
	engine := newSyncPushEngine(&fakeSyncTaskApp{})
	req := httptest.NewRequest(
		http.MethodPost, "/api/sync/push", io.NopCloser(strings.NewReader("{}")),
	)
	req.ContentLength = 10 << 20

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413（body=%s）", rec.Code, rec.Body.String())
	}
}

// postSyncPush 发送一次 /sync/push 请求
func postSyncPush(t *testing.T, engine *gin.Engine, body types.SyncPushReq) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("序列化请求失败: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/sync/push", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

// syncPushResponse 只解出测试关心的字段
type syncPushResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
	Data  struct {
		Results []struct {
			Table   string `json:"table"`
			Outcome string `json:"outcome"`
		} `json:"results"`
	} `json:"data"`
}

func decodeSyncPushRes(t *testing.T, rec *httptest.ResponseRecorder) syncPushResponse {
	t.Helper()
	var res syncPushResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	return res
}
