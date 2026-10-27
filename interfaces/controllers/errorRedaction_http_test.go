package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	userApp "naotodoserver/application/user"
	userDto "naotodoserver/application/user/dto"
	domerr "naotodoserver/domain/errors"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 泄漏探针：内部细节一旦回显，响应体就会包含这些标记
const (
	probePK     = "99199"
	probeTable  = "user_sessions"
	probeSQL    = "SELECT * FROM"
	probePath   = "/opt/naotodo/uploads/avatars/1.png"
	probeGorm   = "record not found"
	probeDupKey = "Duplicate entry"
)

// leakMarkers 响应体一律不得包含的标记（SEC-12）
var leakMarkers = []string{
	probePK, probeTable, probeSQL, probePath, probeGorm, probeDupKey,
	"PRIMARY", "sql:", "Error 1062", "tasks", "projects",
}

// dbDuplicateKeyErr 构造一个"看起来像"真实 MySQL 主键冲突错误（含主键值与表名）
func dbDuplicateKeyErr() error {
	return &mysql.MySQLError{
		Number:  1062,
		Message: "Duplicate entry '" + probePK + "' for key 'PRIMARY'",
	}
}

// fakeErrProjectApp 只替换 Get 的错误
type fakeErrProjectApp struct {
	projectApp.ProjectApp
	getErr error
}

func (f *fakeErrProjectApp) Get(
	_ context.Context, _ int64, _ string,
) (*projectDto.GetProjectRes, error) {
	return nil, f.getErr
}

// fakeErrTaskApp 只替换 UpdateTask / CreateTask 的错误
type fakeErrTaskApp struct {
	taskApp.TaskApp
	updateErr error
	createErr []error
	callIndex int
}

func (f *fakeErrTaskApp) UpdateTask(
	_ context.Context, _ int64, _ string, _ *taskDto.UpdateTaskReq,
) error {
	return f.updateErr
}

func (f *fakeErrTaskApp) CreateTask(
	_ context.Context, _ int64, _ *taskDto.CreateTaskReq,
) (*taskDto.GetTaskRes, domaintypes.UpsertResult, error) {
	if len(f.createErr) == 0 {
		return &taskDto.GetTaskRes{Id: "1"}, domaintypes.UpsertResult{}, nil
	}
	err := f.createErr[f.callIndex%len(f.createErr)]
	f.callIndex++
	return nil, domaintypes.UpsertResult{}, err
}

// fakeErrUserApp 只替换 UpdateAvatar 的错误
type fakeErrUserApp struct {
	userApp.UserApp
	avatarErr error
}

func (f *fakeErrUserApp) UpdateAvatar(
	_ context.Context, _ int64, _ userDto.UpdateAvatarInput,
) (*userDto.UpdateAvatarOutput, error) {
	return nil, f.avatarErr
}

// newProjectGetEngine 挂载真实 ProjectController.GetProject（注入 userId=42）
func newProjectGetEngine(app projectApp.ProjectApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(controllersRequestIdForTest())
	ctrl := controllers.NewProjectController(app)
	engine.GET("/api/projects/:projectId", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.GetProject(ctx)
	})
	return engine
}

// controllersRequestIdForTest 模拟线上 RequestId 中间件（便于断言日志关联）
func controllersRequestIdForTest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetRequestId(ctx.Request.Context(), "test-request-id"))
		ctx.Next()
	}
}

// captureCtlLogs 捕获结构化日志
func captureCtlLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logging.Logger = logger
	t.Cleanup(func() { logging.Logger = previous })
	return buf
}

// getProject 发一次 GET 清单请求
func getProject(t *testing.T, engine *gin.Engine, projectId string) []byte {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectId, nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec.Body.Bytes()
}

// assertNoLeak 断言响应不含任何内部细节标记
func assertNoLeak(t *testing.T, body []byte) {
	t.Helper()
	text := string(body)
	for _, marker := range leakMarkers {
		if strings.Contains(text, marker) {
			t.Fatalf("响应体泄漏内部细节 %q: %s", marker, text)
		}
	}
}

// TestInfraErrorRedactedAndLogged (c) 基础设施错误 ⇒ 通用文案 + 服务端日志（含 request_id）
func TestInfraErrorRedactedAndLogged(t *testing.T) {
	logs := captureCtlLogs(t)
	engine := newProjectGetEngine(&fakeErrProjectApp{getErr: dbDuplicateKeyErr()})

	body := getProject(t, engine, "123")
	assertNoLeak(t, body)

	var res types.ResponseData
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, body)
	}
	if res.Code != 20002 || res.Message != "获取清单失败" {
		t.Fatalf("code/message = %d/%q, want 20002/获取清单失败", res.Code, res.Message)
	}
	if res.Error != "服务暂时不可用，请稍后重试" {
		t.Fatalf("error = %q, want 通用文案", res.Error)
	}

	logText := logs.String()
	for _, want := range []string{"api.error", "test-request-id", "20002", probeDupKey} {
		if !strings.Contains(logText, want) {
			t.Fatalf("服务端日志缺少 %q: %s", want, logText)
		}
	}
}

// TestNotFoundThreeStatesIdenticalBytes (b) 不存在 / 他人 / 已删除 ⇒ 响应逐字节相同
func TestNotFoundThreeStatesIdenticalBytes(t *testing.T) {
	captureCtlLogs(t)
	engine := newProjectGetEngine(&fakeErrProjectApp{getErr: domerr.ErrProjectNotFound})
	otherTenant := getProject(t, engine, "1001")
	notExist := getProject(t, engine, "1002")
	deleted := getProject(t, engine, "1003")

	if !bytes.Equal(otherTenant, notExist) || !bytes.Equal(otherTenant, deleted) {
		t.Fatalf("三态响应必须逐字节相同:\n 他人=%s\n 不存在=%s\n 已删除=%s",
			otherTenant, notExist, deleted)
	}
	assertNoLeak(t, otherTenant)
	var res types.ResponseData
	if err := json.Unmarshal(otherTenant, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if res.Error != "记录不存在" {
		t.Fatalf("error = %q, want 记录不存在", res.Error)
	}
}

// TestGMPRecordNotFoundRedacted 读路径透传的 gorm.ErrRecordNotFound（含 %w 包装）
// 同样归一为「记录不存在」，不透出 ORM 原文
func TestGMPRecordNotFoundRedacted(t *testing.T) {
	captureCtlLogs(t)
	cases := map[string]error{
		"直接返回":   gorm.ErrRecordNotFound,
		"被应用层包装": fmt.Errorf("project.Get: %w", gorm.ErrRecordNotFound),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			engine := newProjectGetEngine(&fakeErrProjectApp{getErr: err})
			body := getProject(t, engine, "1004")
			assertNoLeak(t, body)
			if !strings.Contains(string(body), "记录不存在") {
				t.Fatalf("record not found 应归一为「记录不存在」: %s", body)
			}
		})
	}
}

// TestValidationTextUnchanged (a) 可读校验文案保持不变
func TestValidationTextUnchanged(t *testing.T) {
	captureCtlLogs(t)
	engine := newProjectGetEngine(&fakeErrProjectApp{getErr: domerr.ErrSortIdOverflow})
	body := getProject(t, engine, "1005")
	if !strings.Contains(string(body), domerr.ErrSortIdOverflow.Error()) {
		t.Fatalf("合法校验文案被改写: %s（want 含 %q）", body, domerr.ErrSortIdOverflow.Error())
	}
}

// TestBinderErrorRedacted 参数绑定失败 ⇒ 固定「参数错误」，不回显解析细节
func TestBinderErrorRedacted(t *testing.T) {
	captureCtlLogs(t)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewTaskController(&fakeErrTaskApp{})
	engine.PUT("/api/tasks/:taskId", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.UpdateTask(ctx)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/tasks/1", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	var res types.ResponseData
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	if res.Code != 40021 || res.Message != "参数错误" {
		t.Fatalf("code/message = %d/%q, want 40021/参数错误", res.Code, res.Message)
	}
	if res.Error != "" {
		t.Fatalf("binder 失败不应回显细节，实际 error=%q", res.Error)
	}
	if strings.Contains(rec.Body.String(), "invalid character") {
		t.Fatalf("响应回显了 binder 细节: %s", rec.Body.String())
	}
}

// TestSyncItemErrorWhitelist sync 逐条 Error 只允许 invalid/conflict/server_error
func TestSyncItemErrorWhitelist(t *testing.T) {
	captureCtlLogs(t)
	fake := &fakeErrTaskApp{createErr: []error{
		domerr.ErrIDConflict,
		domerr.Publicf("最多拥有 8 个附件"),
		dbDuplicateKeyErr(),
	}}
	engine := newSyncPushEngineForErrors(fake)

	payload := types.SyncPushReq{Tasks: []types.SyncTaskPushItem{
		syncTaskPushItem("冲突"), syncTaskPushItem("非法"), syncTaskPushItem("服务端错误"),
	}}
	rec := postSyncPushJSON(t, engine, payload)
	// sync 响应按契约携带 table 键（客户端协议的一部分），此处只断言不泄漏 DB/驱动细节
	for _, marker := range []string{probePK, probeSQL, probeDupKey, "PRIMARY", "Error 1062", probeGorm} {
		if strings.Contains(rec.Body.String(), marker) {
			t.Fatalf("逐条结果泄漏内部细节 %q: %s", marker, rec.Body.String())
		}
	}

	var res struct {
		Data types.SyncPushRes `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, rec.Body.String())
	}
	if len(res.Data.Results) != 3 {
		t.Fatalf("逐条结果数 = %d, want 3（body=%s）", len(res.Data.Results), rec.Body.String())
	}
	wantErrors := []string{"conflict", "invalid", "server_error"}
	wantOutcomes := []string{"conflict", "error", "error"}
	for i := range res.Data.Results {
		if res.Data.Results[i].Error != wantErrors[i] {
			t.Fatalf("第 %d 条 Error = %q, want %q", i+1, res.Data.Results[i].Error, wantErrors[i])
		}
		if res.Data.Results[i].Outcome != wantOutcomes[i] {
			t.Fatalf("第 %d 条 Outcome = %q, want %q（语义不变）",
				i+1, res.Data.Results[i].Outcome, wantOutcomes[i])
		}
	}
}

// TestAvatarSaveErrorRedacted 头像保存失败 ⇒ 统一文案，不回显绝对路径
func TestAvatarSaveErrorRedacted(t *testing.T) {
	captureCtlLogs(t)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewUserController(&fakeErrUserApp{
		avatarErr: domerr.Publicf("头像保存失败，请稍后重试"),
	})
	engine.PUT("/api/user/avatar", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.UpdateUserAvatar(ctx)
	})

	payload := `{"avatarURL":"https://example.test/a.png"}`
	req := httptest.NewRequest(http.MethodPut, "/api/user/avatar", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "/opt/") || strings.Contains(body, ".png:") {
		t.Fatalf("响应回显了文件路径: %s", body)
	}
	if !strings.Contains(body, "头像保存失败") {
		t.Fatalf("头像失败文案不符: %s", body)
	}
}

// newSyncPushEngineForErrors 与线上同构的最小 sync 引擎（仅 tasks，app 层用 fake）
func newSyncPushEngineForErrors(fake *fakeErrTaskApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewSyncController(fake, nil, nil, nil, nil, nil)
	engine.POST("/api/sync/push", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.Push(ctx)
	})
	return engine
}

// postSyncPushJSON 发送同步推送请求
func postSyncPushJSON(t *testing.T, engine *gin.Engine, payload types.SyncPushReq) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("序列化请求失败: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/sync/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}
