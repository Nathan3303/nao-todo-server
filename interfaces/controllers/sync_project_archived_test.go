// T322 单测：/sync/push 清单可空时间 archivedAt 三态（absent / null / "" / 值）
//
// 缺陷背景（DEF-42，源于 T318/T319 发现）：`SyncProjectPushItem` 内嵌 `CreateProjectReq` 无
// `archivedAt` 字段，且项目 upsert map 明确不触碰 `archived_at` ⇒ 清单归档/取消归档无法经
// `/sync/push` 传播（只有 REST `PUT /projects/[un]archive/:id` 会级联）。
//
// 修复口径（本单）：
//   - absent ⇒ 不写列（⛔ 不得当清空：旧客户端与 REST create 行为必须不变）
//   - null   ⇒ 显式清空（写 NULL）
//   - ""     ⇒ 显式清空（维持既有约定）
//   - 值     ⇒ 写入
//
// ⛔ 服务端不级联：清单下任务的归档态由客户端逐任务推送完成。
package controllers

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	"naotodoserver/domain/project/entities"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/interfaces/types"
)

// projectSyncPayload 构造单条 sync 清单推送 JSON（可选附加 archivedAt）。
func projectSyncPayload(archived json.RawMessage) string {
	body := `{"projects":[{"id":"8001","name":"清单","description":""`
	if archived != nil {
		body += `,"archivedAt":` + string(archived)
	}
	return body + `}]}`
}

func bindSyncProjectItem(t *testing.T, archived json.RawMessage) types.SyncProjectPushItem {
	t.Helper()
	var req types.SyncPushReq
	if err := json.Unmarshal([]byte(projectSyncPayload(archived)), &req); err != nil {
		t.Fatalf("绑定 sync push: %v", err)
	}
	if len(req.Projects) != 1 {
		t.Fatalf("projects 条数 = %d, want 1", len(req.Projects))
	}
	return req.Projects[0]
}

// TestSyncProjectArchivedAtTriStateMatrix 四态：JSON 绑定 → 应用层 VO 的三态语义。
func TestSyncProjectArchivedAtTriStateMatrix(t *testing.T) {
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
			item := bindSyncProjectItem(t, tc.archived)
			appReq := toCreateProjectInputFromSync(&item)
			vo, err := projectApp.CreateProjectReqToValueObject(1001, appReq)
			if err != nil {
				t.Fatalf("CreateProjectReqToValueObject: %v", err)
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

// TestSyncProjectArchivedAtAbsentIsNotClear 回归锁：缺省（旧客户端 / 未传键）⇒ 应用层入参为 nil，
// 即「不写列」；⛔ 不得被当作清空（否则旧客户端推送会把服务端归档态抹掉）。
func TestSyncProjectArchivedAtAbsentIsNotClear(t *testing.T) {
	item := bindSyncProjectItem(t, nil)
	if item.ArchivedAt.Present {
		t.Fatalf("absent 时三态 Present 应为 false")
	}
	appReq := toCreateProjectInputFromSync(&item)
	if appReq.ArchivedAt != nil {
		t.Fatalf("absent 映射出非 nil *string = %q（会被当作清空 ⇒ 回归）", *appReq.ArchivedAt)
	}
	vo, err := projectApp.CreateProjectReqToValueObject(1001, appReq)
	if err != nil {
		t.Fatalf("CreateProjectReqToValueObject: %v", err)
	}
	if vo.ArchivedAt.Valid {
		t.Fatalf("absent 不得写 archived_at（Valid 应为 false），got %+v", vo.ArchivedAt)
	}
}

// TestCreateProjectRestArchivedAtUnchanged REST create 契约回归：共享 CreateProjectReq 无
// archivedAt 字段 ⇒ 请求体里即便出现 `archivedAt` 也被丢弃，应用层入参恒为 nil ⇒ 不写列；
// 且应用层 sync 专用字段 `json:"-"` 不暴露注入面。
func TestCreateProjectRestArchivedAtUnchanged(t *testing.T) {
	const payload = `{"name":"清单","description":"","archivedAt":null}`
	var restReq types.CreateProjectReq
	if err := json.Unmarshal([]byte(payload), &restReq); err != nil {
		t.Fatalf("绑定 create REST 载荷: %v", err)
	}
	appReq := toCreateProjectInput(&restReq)
	if appReq.ArchivedAt != nil {
		t.Fatalf("REST create 入参 ArchivedAt 应为 nil（无该字段），got %q", *appReq.ArchivedAt)
	}
	vo, err := projectApp.CreateProjectReqToValueObject(1001, appReq)
	if err != nil {
		t.Fatalf("CreateProjectReqToValueObject: %v", err)
	}
	if vo.ArchivedAt.Valid {
		t.Fatalf("REST create 不得写 archived_at，got %+v", vo.ArchivedAt)
	}
	if _, ok := jsonTagSet(reflect.TypeOf(projectDto.CreateProjectReq{}))["archivedAt"]; ok {
		t.Fatalf("应用层 CreateProjectReq 的 ArchivedAt 不得暴露 JSON 注入面（应 json:\"-\"）")
	}
}

// TestSyncProjectPullOutputArchivedAtUnchanged pull 出参回归：未归档清单仍输出 ""。
func TestSyncProjectPullOutputArchivedAtUnchanged(t *testing.T) {
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	unarchived := projectApp.ProjectEntityToGetRes(&entities.Project{
		ArchivedAt: domaintypes.NewNullableTimeNull(),
	})
	if unarchived.ArchivedAt != "" {
		t.Fatalf("未归档清单 archivedAt = %q, want \"\"", unarchived.ArchivedAt)
	}
	archived := projectApp.ProjectEntityToGetRes(&entities.Project{
		ArchivedAt: domaintypes.NewNullableTimeByTime(at),
	})
	if archived.ArchivedAt != at.Format(time.RFC3339) {
		t.Fatalf("已归档清单 archivedAt = %q, want %q", archived.ArchivedAt, at.Format(time.RFC3339))
	}
}
