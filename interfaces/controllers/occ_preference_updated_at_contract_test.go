// T163 契约测试：偏好推送响应 additive 回传服务端权威 updatedAt。
//
// 覆盖 ADR §9.1/§9.4 的偏好回传缺口（F8）：
//   - POST /projects/:projectId/preference 响应 Data 含 projectId + updatedAt（与 GET 同格式）
//   - PUT /user/config 响应 Data 含 updatedAt（与 GET 同格式）
//
// updatedAt 均由应用层「保存后回读」产生，与 GET 走同一转换器（RFC3339Milli），
// 故客户端可直接做字符串比较 / 落 per-row OCC base。
package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	userApp "naotodoserver/application/user"
	userDto "naotodoserver/application/user/dto"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// fakeProjectAppForPreference 仅覆盖 SavePreference；其余方法由内嵌接口占位
type fakeProjectAppForPreference struct {
	projectApp.ProjectApp
	res *projectDto.GetProjectPreferenceRes
	err error
}

func (f *fakeProjectAppForPreference) SavePreference(
	_ context.Context, _ int64, _ string, _ *projectDto.UpdateProjectPreferenceReq,
) (*projectDto.GetProjectPreferenceRes, error) {
	return f.res, f.err
}

// fakeUserAppForConfig 仅覆盖 UpdateConfig；其余方法由内嵌接口占位
type fakeUserAppForConfig struct {
	userApp.UserApp
	output *userDto.GetConfigOutput
	err    error
}

func (f *fakeUserAppForConfig) UpdateConfig(
	_ context.Context, _ int64, _ userDto.UpdateConfigInput,
) (*userDto.GetConfigOutput, error) {
	return f.output, f.err
}

// decodeData 解析成功响应信封中的 data 对象
func decodeData(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("HTTP 状态 = %d, body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	return resp.Data
}

// assertRFC3339Milli 断言时间串可被 RFC3339Milli 解析（与 GET 同格式）
func assertRFC3339Milli(t *testing.T, name, value string) {
	t.Helper()
	if value == "" {
		t.Fatalf("%s 不应为空", name)
	}
	if _, err := time.Parse(idutil.RFC3339Milli, value); err != nil {
		t.Fatalf("%s = %q 不是 RFC3339Milli（与 GET 格式不一致）: %v", name, value, err)
	}
}

// TestSaveProjectPreference_ResponseHasUpdatedAt POST 偏好响应 additive 回传 updatedAt。
func TestSaveProjectPreference_ResponseHasUpdatedAt(t *testing.T) {
	const wantUpdatedAt = "2026-09-24T12:00:00.123Z"
	app := &fakeProjectAppForPreference{res: &projectDto.GetProjectPreferenceRes{
		ProjectId: "88001",
		UpdatedAt: wantUpdatedAt,
	}}
	c := NewProjectController(app)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/projects/88001/preference",
		strings.NewReader(`{"viewType":"table","getTasksOptions":"{}","columns":"[]"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(iCtx.SetUserId(req.Context(), 1001))
	gc.Request = req
	gc.Params = gin.Params{{Key: "projectId", Value: "88001"}}
	c.SaveProjectPreference(gc)

	data := decodeData(t, w)
	if got := data["projectId"]; got != "88001" {
		t.Fatalf("projectId = %v, want 88001（既有字段不得丢失）", got)
	}
	updatedAt, _ := data["updatedAt"].(string)
	if updatedAt != wantUpdatedAt {
		t.Fatalf("updatedAt = %q, want %q", updatedAt, wantUpdatedAt)
	}
	assertRFC3339Milli(t, "updatedAt", updatedAt)

	// 与 GET 同一转换器 ⇒ 同格式（同源保证）
	getRes := toGetProjectPreferenceRes(app.res)
	if getRes.UpdatedAt != updatedAt {
		t.Fatalf("POST updatedAt = %q, GET updatedAt = %q（应同源同格式）", updatedAt, getRes.UpdatedAt)
	}
}

// TestUpdateUserConfig_ResponseHasUpdatedAt PUT /user/config 响应 additive 回传 updatedAt。
func TestUpdateUserConfig_ResponseHasUpdatedAt(t *testing.T) {
	const wantUpdatedAt = "2026-09-24T12:34:56.789Z"
	app := &fakeUserAppForConfig{output: &userDto.GetConfigOutput{UpdatedAt: wantUpdatedAt}}
	c := NewUserController(app)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/user/config",
		strings.NewReader(`{"preferences":{"version":1}}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(iCtx.SetUserId(req.Context(), 1001))
	gc.Request = req
	c.UpdateUserConfig(gc)

	data := decodeData(t, w)
	updatedAt, _ := data["updatedAt"].(string)
	if updatedAt != wantUpdatedAt {
		t.Fatalf("updatedAt = %q, want %q", updatedAt, wantUpdatedAt)
	}
	assertRFC3339Milli(t, "updatedAt", updatedAt)
	// 与 GET 同一转换器（ConfigEntity2Res ⇒ FormatTimeMilli）⇒ 同格式（同源保证）
	if got := toGetConfigRes(app.output).UpdatedAt; got != updatedAt {
		t.Fatalf("PUT updatedAt = %q, GET updatedAt = %q（应同源同格式）", updatedAt, got)
	}
}

// 断言响应类型确实带 updatedAt 字段（JSON tag 契约锁定，防被误删）
func TestPreferenceUpdatedAt_ResponseTypesDeclared(t *testing.T) {
	for name, typ := range map[string]reflect.Type{
		"SaveProjectPreferenceRes": reflect.TypeOf(types.SaveProjectPreferenceRes{}),
		"UpdateUserConfigRes":      reflect.TypeOf(types.UpdateUserConfigRes{}),
	} {
		if _, ok := jsonTagSet(typ)["updatedAt"]; !ok {
			t.Fatalf("%s 缺少 updatedAt 字段", name)
		}
	}
}
