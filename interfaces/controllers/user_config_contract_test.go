// TASK-26 / T130 服务端契约（接口层）：用户配置 preferences blob + 服务端权威 updatedAt。
//
// 覆盖 qa 用例：
//   - CT-01 PUT /user/config 接受全量快照 preferences（推送时装配）
//   - CT-03 GET /user/config 出参含 updatedAt（服务端权威时间）
//   - CT-04 UpdateUserConfigReq 可选化（缺失 appearance 不再 400）
//   - CT-05 preferences 服务端不解析（哑存储，未知键原样透传）
//   - CT-06 appearance 独立不动（仅 appearance / 仅 preferences / 两者 三态）
package controllers

import (
	"encoding/json"
	"reflect"
	"testing"

	userDto "naotodoserver/application/user/dto"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin/binding"
)

// bindUserConfigReq 走 gin 真实 JSON 绑定（含 validator），复现接口层行为
func bindUserConfigReq(t *testing.T, body string) (types.UpdateUserConfigReq, error) {
	t.Helper()
	var req types.UpdateUserConfigReq
	err := binding.JSON.BindBody([]byte(body), &req)
	return req, err
}

// TestUserConfigContract_GetResHasPreferencesAndUpdatedAt CT-03：
// 响应含 preferences（对象，非字符串）与 updatedAt。
func TestUserConfigContract_GetResHasPreferencesAndUpdatedAt(t *testing.T) {
	updatedAt := "2026-09-23T10:20:30.123Z"
	res := toGetConfigRes(&userDto.GetConfigOutput{
		Appearance:  "dark",
		Preferences: `{"version":1,"asideWidth":280}`,
		UpdatedAt:   updatedAt,
	})
	if res.Appearance != "dark" {
		t.Fatalf("appearance = %q, want dark", res.Appearance)
	}
	if res.UpdatedAt != updatedAt {
		t.Fatalf("updatedAt = %q, want %q", res.UpdatedAt, updatedAt)
	}

	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("序列化响应: %v", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("解析响应: %v", err)
	}
	if _, ok := payload["updatedAt"]; !ok {
		t.Fatalf("响应缺少 updatedAt 字段: %s", raw)
	}
	var preferences map[string]any
	if err := json.Unmarshal(payload["preferences"], &preferences); err != nil {
		t.Fatalf("preferences 应为 JSON 对象而非字符串: %v (%s)", err, raw)
	}
	if preferences["version"] != float64(1) {
		t.Fatalf("preferences 未原样回传: %s", payload["preferences"])
	}
}

// TestUserConfigContract_GetResEmptyPreferencesIsObject CT-05：
// 未设置偏好时返回空对象（客户端以本地为准并回传）。
func TestUserConfigContract_GetResEmptyPreferencesIsObject(t *testing.T) {
	res := toGetConfigRes(&userDto.GetConfigOutput{Appearance: "auto"})
	if string(res.Preferences) != "{}" {
		t.Fatalf("未设置偏好应返回空对象，got %s", res.Preferences)
	}
}

// TestUserConfigContract_UpdateReqAppearanceOptional CT-04 / CT-06：
// 仅 preferences / 仅 appearance / 两者 三种 PUT 均可绑定并正确透传。
func TestUserConfigContract_UpdateReqAppearanceOptional(t *testing.T) {
	t.Run("仅 preferences（缺失 appearance 不再 400）", func(t *testing.T) {
		req, err := bindUserConfigReq(t, `{"preferences":{"version":1,"asideWidth":280}}`)
		if err != nil {
			t.Fatalf("缺失 appearance 应可绑定: %v", err)
		}
		input := toUpdateConfigInput(req)
		if input.Appearance != nil {
			t.Fatalf("未携带 appearance 时应为 nil，got %v", *input.Appearance)
		}
		if input.Preferences == nil || *input.Preferences != `{"version":1,"asideWidth":280}` {
			t.Fatalf("preferences 未透传: %v", input.Preferences)
		}
	})

	t.Run("仅 appearance", func(t *testing.T) {
		req, err := bindUserConfigReq(t, `{"appearance":"dark"}`)
		if err != nil {
			t.Fatalf("绑定 appearance: %v", err)
		}
		input := toUpdateConfigInput(req)
		if input.Appearance == nil || *input.Appearance != "dark" {
			t.Fatalf("appearance 未透传: %v", input.Appearance)
		}
		if input.Preferences != nil {
			t.Fatal("未携带 preferences 时应为 nil（不改该列）")
		}
	})

	t.Run("两者", func(t *testing.T) {
		req, err := bindUserConfigReq(t, `{"appearance":"light","preferences":{"version":1}}`)
		if err != nil {
			t.Fatalf("绑定两者: %v", err)
		}
		input := toUpdateConfigInput(req)
		if input.Appearance == nil || *input.Appearance != "light" {
			t.Fatalf("appearance 未透传: %v", input.Appearance)
		}
		if input.Preferences == nil || *input.Preferences != `{"version":1}` {
			t.Fatalf("preferences 未透传: %v", input.Preferences)
		}
	})
}

// TestUserConfigContract_PreferencesOpaque CT-05 / PS-5：
// 服务端不解析偏好内容：未知键（含未来新增偏好）原样透传，不被丢弃或改写。
func TestUserConfigContract_PreferencesOpaque(t *testing.T) {
	blob := `{"version":1,"unknownFuturePreference":{"nested":[1,2,3]},` +
		`"builtInProjectPreferences":{"all":{"viewType":"kanban"}}}`
	req, err := bindUserConfigReq(t, `{"preferences":`+blob+`}`)
	if err != nil {
		t.Fatalf("绑定: %v", err)
	}
	input := toUpdateConfigInput(req)
	if input.Preferences == nil || *input.Preferences != blob {
		t.Fatalf("偏好快照应原样透传（哑存储）:\n got=%v\nwant=%s", input.Preferences, blob)
	}
}

// TestUserConfigContract_PreferencesNullClears CT-05：
// 显式 null 视为清除（写 NULL），不把 JSON null 存入偏好快照。
func TestUserConfigContract_PreferencesNullClears(t *testing.T) {
	req, err := bindUserConfigReq(t, `{"preferences":null}`)
	if err != nil {
		t.Fatalf("绑定 null: %v", err)
	}
	input := toUpdateConfigInput(req)
	if input.Preferences == nil {
		t.Fatal("显式 null 应表达「清除」而非「不修改」")
	}
	if *input.Preferences != "" {
		t.Fatalf("显式 null 应映射为空串（清除），got %q", *input.Preferences)
	}
}

// TestUserConfigContract_UpdateReqFormBindingSkipsPreferences CT-05 边界：
// preferences 仅走 JSON 绑定（form 标记为 -），避免表单绑定路径处理 RawMessage。
func TestUserConfigContract_UpdateReqFormBindingSkipsPreferences(t *testing.T) {
	field, ok := reflect.TypeOf(types.UpdateUserConfigReq{}).FieldByName("Preferences")
	if !ok {
		t.Fatal("UpdateUserConfigReq 缺少 Preferences 字段")
	}
	if got := field.Tag.Get("form"); got != "-" {
		t.Fatalf("Preferences form tag = %q, want -", got)
	}
	if got := field.Tag.Get("json"); got != "preferences" {
		t.Fatalf("Preferences json tag = %q, want preferences", got)
	}
}
