// 子任务排序（sort_id）接口层透传（ADR 2026-09-13 §9.1 三处补 sortId + sync push 复用同一转换器）
// 客户端重排值必须能经 HTTP/sync 入口进入应用层；`CreateTaskReq` 三处（types/dto/转换器）任一处漏接都会静默丢弃。
package controllers

import (
	"encoding/json"
	"testing"

	"naotodoserver/interfaces/types"
)

// TestSortId_ToCreateTaskReq_MapsSortId 断言 CreateTaskReq.sortId 经接口层转换器透传到应用层入参
func TestSortId_ToCreateTaskReq_MapsSortId(t *testing.T) {
	req := &types.CreateTaskReq{
		Name:     "子任务",
		State:    "todo",
		Priority: "medium",
		SortId:   4000,
	}
	out := toCreateTaskReq(req)
	if out.SortId != 4000 {
		t.Fatalf("toCreateTaskReq 丢弃 sortId：got %d, want 4000", out.SortId)
	}
}

// TestSortId_CreateTaskReqJSONBindingAndZeroSemantics 断言 JSON `sortId` 可绑定，且未携带时为 0（= 未设置，服务端置组末）
func TestSortId_CreateTaskReqJSONBindingAndZeroSemantics(t *testing.T) {
	var withSort types.CreateTaskReq
	if err := json.Unmarshal([]byte(`{"name":"n","state":"todo","priority":"medium","sortId":5000}`), &withSort); err != nil {
		t.Fatalf("绑定 sortId: %v", err)
	}
	if withSort.SortId != 5000 {
		t.Fatalf("JSON sortId 绑定 = %d, want 5000", withSort.SortId)
	}
	if got := toCreateTaskReq(&withSort).SortId; got != 5000 {
		t.Fatalf("JSON 绑定后透传 = %d, want 5000", got)
	}

	var withoutSort types.CreateTaskReq
	if err := json.Unmarshal([]byte(`{"name":"n","state":"todo","priority":"medium"}`), &withoutSort); err != nil {
		t.Fatalf("绑定缺省: %v", err)
	}
	if withoutSort.SortId != 0 {
		t.Fatalf("未携带 sortId 应为 0（未设置语义），got %d", withoutSort.SortId)
	}
	if got := toCreateTaskReq(&withoutSort).SortId; got != 0 {
		t.Fatalf("未携带时透传 = %d, want 0", got)
	}
}

// TestSortId_ToUpdateTaskReq_MapsSortId 断言 PATCH 侧 sortId 透传（全链已通，防回归）
func TestSortId_ToUpdateTaskReq_MapsSortId(t *testing.T) {
	v := uint16(7000)
	req := &types.UpdateTaskReq{SortId: &v}
	out := toUpdateTaskReq(req)
	if out.SortId == nil || *out.SortId != 7000 {
		t.Fatalf("toUpdateTaskReq 丢弃 sortId：got %v, want 7000", out.SortId)
	}
	if toUpdateTaskReq(&types.UpdateTaskReq{}).SortId != nil {
		t.Fatal("UpdateTaskReq.SortId 未携带时应为 nil（不改该列）")
	}
}
