package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userApp "naotodoserver/application/user"
	userDto "naotodoserver/application/user/dto"
	domerr "naotodoserver/domain/errors"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"

	"github.com/gin-gonic/gin"
)

// fakePasswordUserApp 只实现改密路径；其余方法由嵌入接口兜底（本测试不会调用）。
type fakePasswordUserApp struct {
	userApp.UserApp
	calls      int
	lastUserId int64
	lastInput  userDto.UpdatePasswordInput
	err        error
}

func (f *fakePasswordUserApp) UpdatePassword(
	_ context.Context,
	userId int64,
	req userDto.UpdatePasswordInput,
) error {
	f.calls++
	f.lastUserId = userId
	f.lastInput = req
	return f.err
}

// newPasswordEngine 挂载真实控制器，并注入鉴权上下文（user_id + 当次 token）。
func newPasswordEngine(app *fakePasswordUserApp) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controllers.NewUserController(app)
	engine.PUT("/api/user/password", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(
			iCtx.SetToken(iCtx.SetUserId(ctx.Request.Context(), 42), "current-token"),
		)
		ctrl.UpdateUserPassword(ctx)
	})
	return engine
}

type passwordResponse struct {
	Code int `json:"code"`
}

// putPassword 发送改密请求
func putPassword(t *testing.T, engine *gin.Engine, oldPassword, newPassword string) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(map[string]string{
		"oldPassword": oldPassword,
		"newPassword": newPassword,
	})
	if err != nil {
		t.Fatalf("序列化请求失败: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/user/password", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

// TestUpdatePasswordInjectsCurrentToken 改密成功：当次 token 由控制器注入应用层（保留当前设备）
func TestUpdatePasswordInjectsCurrentToken(t *testing.T) {
	app := &fakePasswordUserApp{}
	rec := putPassword(t, newPasswordEngine(app), "old-pass-1", "new-pass-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200（body=%s）", rec.Code, rec.Body.String())
	}
	var res passwordResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if res.Code != 10070 {
		t.Fatalf("code = %d, want 10070", res.Code)
	}
	if app.calls != 1 || app.lastUserId != 42 {
		t.Fatalf("应用层调用 = %d 次 / userId=%d, want 1 次 / 42", app.calls, app.lastUserId)
	}
	if app.lastInput.Token != "current-token" {
		t.Fatalf("应用层收到 token = %q, want current-token", app.lastInput.Token)
	}
	if app.lastInput.OldPassword != "old-pass-1" || app.lastInput.NewPassword != "new-pass-1" {
		t.Fatalf("应用层收到密码不符: %+v", app.lastInput)
	}
}

// TestUpdatePasswordWeakPasswordMapped 口令策略拒绝（SEC-05，由应用层判定）⇒ 10073 + 统一可读文案，
// 且不泄露命中了哪条规则；不撤销会话由应用层顺序保证（见 application/user 单测）。
func TestUpdatePasswordWeakPasswordMapped(t *testing.T) {
	cases := map[string]string{
		"过短":    strings.Repeat("a", 7),
		"过长":    strings.Repeat("a", 65),
		"全同":    strings.Repeat("a", 8),
		"纯数字短":  "12345678",
		"名单命中":  "password",
		"等于邮箱段": "victim",
	}
	for name, newPassword := range cases {
		t.Run(name, func(t *testing.T) {
			app := &fakePasswordUserApp{err: domerr.ErrPasswordTooWeak}
			rec := putPassword(t, newPasswordEngine(app), "old-pass-1", newPassword)

			var res struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("响应不是合法 JSON: %v", err)
			}
			if res.Code != 10073 {
				t.Fatalf("code = %d, want 10073", res.Code)
			}
			if res.Message != "密码长度需 8–64 位且不能过于简单" {
				t.Fatalf("message = %q, want 统一可读文案", res.Message)
			}
			for _, leak := range []string{"弱口令", "名单", "本地部分", "全同", "纯数字"} {
				if strings.Contains(res.Message, leak) {
					t.Fatalf("文案不得泄露规则细节（命中 %q）: %s", leak, res.Message)
				}
			}
			if app.calls != 1 {
				t.Fatalf("控制器应把口令交给应用层判定，实际调用 %d 次", app.calls)
			}
		})
	}
}

// TestUpdatePasswordMissingFieldRejected 参数缺失 ⇒ 10071，且不进入应用层
func TestUpdatePasswordMissingFieldRejected(t *testing.T) {
	app := &fakePasswordUserApp{}
	rec := putPassword(t, newPasswordEngine(app), "", "new-pass-1")

	var res passwordResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("响应不是合法 JSON: %v", err)
	}
	if res.Code != 10071 {
		t.Fatalf("code = %d, want 10071", res.Code)
	}
	if app.calls != 0 {
		t.Fatalf("参数错误时不得调用应用层，实际 %d 次", app.calls)
	}
}
