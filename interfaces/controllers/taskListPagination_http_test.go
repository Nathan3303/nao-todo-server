package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/utils/query"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// fakeTaskListApp 记录列表入参并回显分页（验证「夹取后再下发应用层」）
type fakeTaskListApp struct {
	taskApp.TaskApp
	calls   int
	lastReq *taskDto.ListTaskReq
}

func (f *fakeTaskListApp) ListTask(
	_ context.Context, _ int64, req *taskDto.ListTaskReq,
) (taskDto.ListTaskRes, *taskDto.Pagination, error) {
	f.calls++
	f.lastReq = req
	return taskDto.ListTaskRes{}, &taskDto.Pagination{Page: req.Page, Limit: req.Limit}, nil
}

func newTaskListEngine(fake *fakeTaskListApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewTaskController(fake)
	engine.GET("/api/tasks", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.ListTask(ctx)
	})
	return engine
}

type taskListResponse struct {
	Code       int    `json:"code"`
	Error      string `json:"error"`
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	} `json:"pagination"`
}

// TestListTaskClampsLimit limit=100000 ⇒ 夹取到 500（入参与响应回显一致）
func TestListTaskClampsLimit(t *testing.T) {
	fake := &fakeTaskListApp{}
	rec := httptest.NewRecorder()
	newTaskListEngine(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tasks?limit=100000", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200（body=%s）", rec.Code, rec.Body.String())
	}
	if fake.lastReq == nil || fake.lastReq.Limit != query.MaxLimit {
		t.Fatalf("应用层收到 limit = %v, want %d", fake.lastReq, query.MaxLimit)
	}
	var res taskListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if res.Pagination.Limit != query.MaxLimit {
		t.Fatalf("响应回显 limit = %d, want %d", res.Pagination.Limit, query.MaxLimit)
	}
}

// TestListTaskRejectsDeepPage page=10001 ⇒ 400，且不调用应用层
func TestListTaskRejectsDeepPage(t *testing.T) {
	fake := &fakeTaskListApp{}
	rec := httptest.NewRecorder()
	url := "/api/tasks?page=" + strconv.Itoa(query.MaxPage+1) + "&limit=10"
	newTaskListEngine(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400（body=%s）", rec.Code, rec.Body.String())
	}
	if fake.calls != 0 {
		t.Fatalf("深分页不应触达应用层，实际调用 %d 次", fake.calls)
	}
	var res taskListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if !strings.Contains(res.Error, "page 超出上限") {
		t.Fatalf("error = %q, want 含「page 超出上限」", res.Error)
	}
}

// TestListTaskKeepsLegalPagination 合法分页回归：page=2&limit=10 原样下发
func TestListTaskKeepsLegalPagination(t *testing.T) {
	fake := &fakeTaskListApp{}
	rec := httptest.NewRecorder()
	newTaskListEngine(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tasks?page=2&limit=10", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if fake.lastReq == nil || fake.lastReq.Page != 2 || fake.lastReq.Limit != 10 {
		t.Fatalf("应用层收到 %+v, want page=2 limit=10", fake.lastReq)
	}
}
