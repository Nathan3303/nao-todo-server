package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	userApp "naotodoserver/application/user"
	userDto "naotodoserver/application/user/dto"
	domerr "naotodoserver/domain/errors"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// wrapPrefixPattern 内部 wrap 前缀形态：pkg.Fn: / tk7s.app: 等（F1 断言不得命中）
var wrapPrefixPattern = regexp.MustCompile(`^[a-zA-Z0-9_/]+(\.[A-Za-z0-9_]+)+:`)

// fakeWrappedProjectApp 返回被 fmt.Errorf 包装过的错误（模拟应用层 wrap）
type fakeWrappedProjectApp struct {
	projectApp.ProjectApp
	getErr    error
	createErr error
}

func (f *fakeWrappedProjectApp) Get(
	_ context.Context, _ int64, _ string,
) (*projectDto.GetProjectRes, error) {
	return nil, f.getErr
}

func (f *fakeWrappedProjectApp) Create(
	_ context.Context, _ int64, _ *projectDto.CreateProjectReq,
) (*projectDto.CreateProjectRes, domaintypes.UpsertResult, error) {
	return nil, domaintypes.UpsertResult{}, f.createErr
}

// fakeWrappedUserApp 改密路径返回被包装的可读业务错误
type fakeWrappedUserApp struct {
	userApp.UserApp
	err error
}

func (f *fakeWrappedUserApp) UpdatePassword(
	_ context.Context, _ int64, _ userDto.UpdatePasswordInput,
) error {
	return f.err
}

// newWrappedProjectEngine 挂载真实 ProjectController（Get + Create）
func newWrappedProjectEngine(app projectApp.ProjectApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := controllers.NewProjectController(app)
	engine := gin.New()
	engine.GET("/api/projects/:projectId", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.GetProject(ctx)
	})
	engine.POST("/api/projects", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.CreateProject(ctx)
	})
	return engine
}

// decodeErrRes 解析失败的响应
func decodeErrRes(t *testing.T, body []byte) (int, string, string) {
	t.Helper()
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v（body=%s）", err, body)
	}
	return res.Code, res.Message, res.Error
}

// assertNoWrapPrefix 断言文案不含内部 wrap 前缀
func assertNoWrapPrefix(t *testing.T, label, text string) {
	t.Helper()
	if wrapPrefixPattern.MatchString(text) {
		t.Fatalf("%s 含内部函数前缀: %q", label, text)
	}
	for _, prefix := range []string{"project.Create:", "project.Get:", "tag.Create:", "task.", "user."} {
		if strings.Contains(text, prefix) {
			t.Fatalf("%s 含内部函数前缀 %q: %q", label, prefix, text)
		}
	}
}

// TestIDConflictTextHasNoWrapPrefix F1：被 wrap 的 ErrIDConflict ⇒ 对外只给规范文案
func TestIDConflictTextHasNoWrapPrefix(t *testing.T) {
	captureCtlLogs(t)
	wrapped := fmt.Errorf("project.Create: %w", domerr.ErrIDConflict)

	t.Run("读取路径", func(t *testing.T) {
		engine := newWrappedProjectEngine(&fakeWrappedProjectApp{getErr: wrapped})
		req := httptest.NewRequest(http.MethodGet, "/api/projects/1", nil)
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		code, _, errText := decodeErrRes(t, rec.Body.Bytes())
		if code != 20002 {
			t.Fatalf("code = %d, want 20002", code)
		}
		if errText != domerr.ErrIDConflict.Error() {
			t.Fatalf("error = %q, want 规范文案 %q", errText, domerr.ErrIDConflict.Error())
		}
		assertNoWrapPrefix(t, "error", errText)
	})

	t.Run("创建路径", func(t *testing.T) {
		engine := newWrappedProjectEngine(&fakeWrappedProjectApp{
			createErr: fmt.Errorf("project.Create: %w", domerr.ErrIDConflict),
		})
		req := httptest.NewRequest(
			http.MethodPost, "/api/projects", strings.NewReader(`{"name":"x"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, req)

		code, _, errText := decodeErrRes(t, rec.Body.Bytes())
		if code != 20012 {
			t.Fatalf("code = %d, want 20012", code)
		}
		if errText != domerr.ErrIDConflict.Error() {
			t.Fatalf("error = %q, want 规范文案", errText)
		}
		assertNoWrapPrefix(t, "error", errText)
	})
}

// TestPublicMessageTextHasNoWrapPrefix F1：原本把 err.Error() 当 Message 的端点同样去前缀
func TestPublicMessageTextHasNoWrapPrefix(t *testing.T) {
	captureCtlLogs(t)
	gin.SetMode(gin.TestMode)
	app := &fakeWrappedUserApp{
		err: fmt.Errorf("user.UpdatePassword: %w", domerr.Publicf("旧密码错误")),
	}
	ctrl := controllers.NewUserController(app)
	engine := gin.New()
	engine.PUT("/api/user/password", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.UpdateUserPassword(ctx)
	})

	req := httptest.NewRequest(
		http.MethodPut, "/api/user/password",
		strings.NewReader(`{"oldPassword":"o","newPassword":"brand-new-pass-9"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	code, message, _ := decodeErrRes(t, rec.Body.Bytes())
	if code != 10075 {
		t.Fatalf("code = %d, want 10075", code)
	}
	if message != "旧密码错误" {
		t.Fatalf("message = %q, want 规范文案「旧密码错误」", message)
	}
	assertNoWrapPrefix(t, "message", message)
}
