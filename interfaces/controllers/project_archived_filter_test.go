// DP-1=(b) 服务端回归：清单列表 isArchived 过滤参数解析 ——
// 未传（空串）与显式 "false" 同为零值 false（同路径，无隐式分叉），仅 "true" 返回已归档。
package controllers

import "testing"

func TestParseIsArchivedFilter(t *testing.T) {
	// 未传（空串）与显式 false 必须同结果
	unset, err := parseIsArchivedFilter("")
	if err != nil {
		t.Fatalf("未传 isArchived 解析失败: %v", err)
	}
	explicitFalse, err := parseIsArchivedFilter("false")
	if err != nil {
		t.Fatalf("显式 false 解析失败: %v", err)
	}
	if unset != explicitFalse {
		t.Fatalf("未传与显式 false 必须同结果，实际 unset=%v explicit=%v", unset, explicitFalse)
	}
	if unset {
		t.Fatalf("未传 isArchived 应为 false（默认排除归档），实际 %v", unset)
	}

	explicitTrue, err := parseIsArchivedFilter("true")
	if err != nil {
		t.Fatalf("显式 true 解析失败: %v", err)
	}
	if !explicitTrue {
		t.Fatalf("显式 true 应为 true，实际 %v", explicitTrue)
	}

	if _, err := parseIsArchivedFilter("yes"); err == nil {
		t.Fatal("非法 isArchived 取值应返回错误")
	}
}
