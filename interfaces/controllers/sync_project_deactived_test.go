// T466 单测：/sync/push 清单停用时间 deactivedAt 三态（absent / null / "" / 值）。
//
// 缺陷背景（DEF-46）：`SyncProjectPushItem` 内嵌 `CreateProjectReq` 无 `deactivedAt`
// 字段，且项目 upsert map 不触碰 `deactived_at` ⇒ 清单「仅停用不删除」无法经 /sync/push
// 传播（服务端写 `deactived_at` 的唯一落点是 UpdateState：仅 REST Delete/Restore 走它）。
//
// 修复口径（与 archivedAt / T322 / DEF-42 同一套 NullableString 真三态，不新造第二套）：
//   - absent ⇒ 不写列（⛔ 不得当清空：旧客户端与 REST create 行为必须不变）
//   - null   ⇒ 显式清空（写 NULL）
//   - ""     ⇒ 显式清空（维持既有约定）
//   - 值     ⇒ 写入
package controllers

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	"naotodoserver/interfaces/types"
)

// projectDeactivedSyncPayload 构造单条 sync 清单推送 JSON（可选附加 deactivedAt）。
func projectDeactivedSyncPayload(deactived json.RawMessage) string {
	body := `{"projects":[{"id":"8001","name":"清单","description":""`
	if deactived != nil {
		body += `,"deactivedAt":` + string(deactived)
	}
	return body + `}]}`
}

func bindSyncProjectDeactivedItem(t *testing.T, deactived json.RawMessage) types.SyncProjectPushItem {
	t.Helper()
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(projectDeactivedSyncPayload(deactived)), &req); err != nil {
		t.Fatalf("绑定 sync push: %v", err)
	}
	if len(req.Projects) != 1 {
		t.Fatalf("projects 条数 = %d, want 1", len(req.Projects))
	}
	return req.Projects[0]
}

// TestSyncProjectDeactivedAtTriStateMatrix 四态：JSON 绑定 → 应用层 VO 的三态语义。
func TestSyncProjectDeactivedAtTriStateMatrix(t *testing.T) {
	const value = "2026-01-01T00:00:00Z"
	cases := []struct {
		name      string
		deactived json.RawMessage
		voValid   bool
		voNull    bool
		voTimeStr string
	}{
		{name: "absent", deactived: nil, voValid: false, voNull: true},
		{name: "null", deactived: json.RawMessage(`null`), voValid: true, voNull: true},
		{name: "empty", deactived: json.RawMessage(`""`), voValid: true, voNull: true},
		{name: "value", deactived: json.RawMessage(`"` + value + `"`), voValid: true, voNull: false, voTimeStr: value},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := bindSyncProjectDeactivedItem(t, tc.deactived)
			appReq := toCreateProjectInputFromSync(&item)
			vo, err := projectApp.CreateProjectReqToValueObject(1001, appReq)
			if err != nil {
				t.Fatalf("CreateProjectReqToValueObject: %v", err)
			}
			if vo.DeactivedAt.Valid != tc.voValid || vo.DeactivedAt.IsNull != tc.voNull {
				t.Fatalf("VO.DeactivedAt = {Valid:%v IsNull:%v}, want {Valid:%v IsNull:%v}",
					vo.DeactivedAt.Valid, vo.DeactivedAt.IsNull, tc.voValid, tc.voNull)
			}
			if tc.voTimeStr != "" {
				want, _ := time.Parse(time.RFC3339, tc.voTimeStr)
				if !vo.DeactivedAt.Time.Equal(want) {
					t.Fatalf("VO.DeactivedAt.Time = %v, want %v", vo.DeactivedAt.Time, want)
				}
			}
		})
	}
}

// TestSyncProjectDeactivedAtAbsentIsNotClear 回归锁：缺省（旧客户端 / 未传键）⇒ 应用层入参
// 为 nil，即「不写列」；⛔ 不得被当作清空（否则旧客户端推送会把服务端停用态抹掉）。
func TestSyncProjectDeactivedAtAbsentIsNotClear(t *testing.T) {
	item := bindSyncProjectDeactivedItem(t, nil)
	if item.DeactivedAt.Present {
		t.Fatalf("absent 时三态 Present 应为 false")
	}
	appReq := toCreateProjectInputFromSync(&item)
	if appReq.DeactivedAt != nil {
		t.Fatalf("absent 映射出非 nil *string = %q（会被当作清空 ⇒ 回归）", *appReq.DeactivedAt)
	}
	vo, err := projectApp.CreateProjectReqToValueObject(1001, appReq)
	if err != nil {
		t.Fatalf("CreateProjectReqToValueObject: %v", err)
	}
	if vo.DeactivedAt.Valid {
		t.Fatalf("absent 不得写 deactived_at（Valid 应为 false），got %+v", vo.DeactivedAt)
	}
}

// TestCreateProjectRestDeactivedAtUnchanged REST create 契约回归：共享 CreateProjectReq 无
// deactivedAt 字段 ⇒ 请求体里即便出现 `deactivedAt` 也被丢弃，应用层入参恒为 nil ⇒ 不写列；
// 且应用层 sync 专用字段 `json:"-"` 不暴露注入面。
func TestCreateProjectRestDeactivedAtUnchanged(t *testing.T) {
	const payload = `{"name":"清单","description":"","deactivedAt":null}`
	var restReq types.CreateProjectReq
	if err := json.Unmarshal([]byte(payload), &restReq); err != nil {
		t.Fatalf("绑定 create REST 载荷: %v", err)
	}
	appReq := toCreateProjectInput(&restReq)
	if appReq.DeactivedAt != nil {
		t.Fatalf("REST create 入参 DeactivedAt 应为 nil（无该字段），got %q", *appReq.DeactivedAt)
	}
	vo, err := projectApp.CreateProjectReqToValueObject(1001, appReq)
	if err != nil {
		t.Fatalf("CreateProjectReqToValueObject: %v", err)
	}
	if vo.DeactivedAt.Valid {
		t.Fatalf("REST create 不得写 deactived_at，got %+v", vo.DeactivedAt)
	}
	if _, ok := jsonTagSet(reflect.TypeOf(projectDto.CreateProjectReq{}))["deactivedAt"]; ok {
		t.Fatalf("应用层 CreateProjectReq 的 DeactivedAt 不得暴露 JSON 注入面（应 json:\"-\"）")
	}
}
