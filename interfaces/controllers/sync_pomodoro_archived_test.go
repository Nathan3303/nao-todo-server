// T466 单测：/sync/push 常用番茄工作归档时间 archivedAt 三态（absent / null / "" / 值）。
//
// 缺陷背景（DEF-43）：客户端 pomodoros 载荷含 `archivedAt` 键（与 DEF-42 同源的
// 「字段缺失族」），而服务端 `CreatePomodoroReq` 无该承载 ⇒ 静默丢弃。
//
// 修复口径（与清单 archivedAt / deactivedAt 同一套 NullableString 真三态）：
//   - absent ⇒ 不写列（旧客户端与 REST create 行为不变）
//   - null   ⇒ 显式清空（写 NULL）
//   - ""     ⇒ 显式清空（维持既有约定）
//   - 值     ⇒ 写入
package controllers

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroDto "naotodoserver/application/pomodoro/dto"
	"naotodoserver/interfaces/types"
)

// pomodoroSyncPayload 构造单条 sync 常用番茄工作推送 JSON（可选附加 archivedAt）。
func pomodoroSyncPayload(archived json.RawMessage) string {
	body := `{"pomodoros":[{"type":1,"name":"番茄","duration":1500`
	if archived != nil {
		body += `,"archivedAt":` + string(archived)
	}
	return body + `}]}`
}

func bindSyncPomodoroItem(t *testing.T, archived json.RawMessage) types.SyncPomodoroPushItem {
	t.Helper()
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(pomodoroSyncPayload(archived)), &req); err != nil {
		t.Fatalf("绑定 sync push: %v", err)
	}
	if len(req.Pomodoros) != 1 {
		t.Fatalf("pomodoros 条数 = %d, want 1", len(req.Pomodoros))
	}
	return req.Pomodoros[0]
}

// TestSyncPomodoroArchivedAtTriStateMatrix 四态：JSON 绑定 → 应用层 VO 的三态语义。
func TestSyncPomodoroArchivedAtTriStateMatrix(t *testing.T) {
	const value = "2026-01-01T00:00:00Z"
	cases := []struct {
		name      string
		archived  json.RawMessage
		voValid   bool
		voNull    bool
		voTimeStr string
	}{
		{name: "absent", archived: nil, voValid: false, voNull: true},
		{name: "null", archived: json.RawMessage(`null`), voValid: true, voNull: true},
		{name: "empty", archived: json.RawMessage(`""`), voValid: true, voNull: true},
		{name: "value", archived: json.RawMessage(`"` + value + `"`), voValid: true, voNull: false, voTimeStr: value},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := bindSyncPomodoroItem(t, tc.archived)
			appReq := toCreatePomodoroInputFromSync(&item)
			vo, err := pomodoroApp.CreatePomodoroReqToVO(1001, appReq)
			if err != nil {
				t.Fatalf("CreatePomodoroReqToVO: %v", err)
			}
			if vo.ArchivedAt.Valid != tc.voValid || vo.ArchivedAt.IsNull != tc.voNull {
				t.Fatalf("VO.ArchivedAt = {Valid:%v IsNull:%v}, want {Valid:%v IsNull:%v}",
					vo.ArchivedAt.Valid, vo.ArchivedAt.IsNull, tc.voValid, tc.voNull)
			}
			if tc.voTimeStr != "" {
				want, _ := time.Parse(time.RFC3339, tc.voTimeStr)
				if !vo.ArchivedAt.Time.Equal(want) {
					t.Fatalf("VO.ArchivedAt.Time = %v, want %v", vo.ArchivedAt.Time, want)
				}
			}
		})
	}
}

// TestSyncPomodoroArchivedAtAbsentIsNotClear 回归锁：缺省 ⇒ 不写列，不得当清空。
func TestSyncPomodoroArchivedAtAbsentIsNotClear(t *testing.T) {
	item := bindSyncPomodoroItem(t, nil)
	if item.ArchivedAt.Present {
		t.Fatalf("absent 时三态 Present 应为 false")
	}
	appReq := toCreatePomodoroInputFromSync(&item)
	if appReq.ArchivedAt != nil {
		t.Fatalf("absent 映射出非 nil *string = %q（会被当作清空 ⇒ 回归）", *appReq.ArchivedAt)
	}
	vo, err := pomodoroApp.CreatePomodoroReqToVO(1001, appReq)
	if err != nil {
		t.Fatalf("CreatePomodoroReqToVO: %v", err)
	}
	if vo.ArchivedAt.Valid {
		t.Fatalf("absent 不得写 archived_at（Valid 应为 false），got %+v", vo.ArchivedAt)
	}
}

// TestCreatePomodoroRestArchivedAtUnchanged REST create 契约回归：共享 CreatePomodoroReq 无
// archivedAt 字段；应用层 sync 专用字段 `json:"-"` 不暴露注入面。
func TestCreatePomodoroRestArchivedAtUnchanged(t *testing.T) {
	const payload = `{"type":1,"name":"番茄","duration":1500,"archivedAt":null}`
	var restReq types.CreatePomodoroReq
	if err := json.Unmarshal([]byte(payload), &restReq); err != nil {
		t.Fatalf("绑定 create REST 载荷: %v", err)
	}
	appReq := toCreatePomodoroInput(restReq)
	if appReq.ArchivedAt != nil {
		t.Fatalf("REST create 入参 ArchivedAt 应为 nil（无该字段），got %q", *appReq.ArchivedAt)
	}
	vo, err := pomodoroApp.CreatePomodoroReqToVO(1001, appReq)
	if err != nil {
		t.Fatalf("CreatePomodoroReqToVO: %v", err)
	}
	if vo.ArchivedAt.Valid {
		t.Fatalf("REST create 不得写 archived_at，got %+v", vo.ArchivedAt)
	}
	if _, ok := jsonTagSet(reflect.TypeOf(pomodoroDto.CreatePomodoroReq{}))["archivedAt"]; ok {
		t.Fatalf("应用层 CreatePomodoroReq 的 ArchivedAt 不得暴露 JSON 注入面（应 json:\"-\"）")
	}
}
