// 领域统计属性联动：计数 server-owned 负向断言（AC10 / ADR B3）
// 计数列由服务端事件联动维护，客户端不得经请求体注入：
// 接口层 req 结构体不存在计数字段 ⇒ JSON 绑定无法注入；经接口层转换器产出的应用层入参同样不含计数。
package controllers

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	projectDto "naotodoserver/application/project/dto"
	taskDto "naotodoserver/application/task/dto"
	"naotodoserver/interfaces/types"
)

// countFieldNames 计数字段名（JSON tag / Go 字段名，小写比较）
var countFieldNames = []string{"checkitemcount", "commentcount", "subtaskcount", "taskcount"}

// assertNoCountField 断言结构体（含嵌入字段）不存在任何计数字段
func assertNoCountField(t *testing.T, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			assertNoCountField(t, f.Type)
			continue
		}
		name := strings.ToLower(f.Name)
		jsonTag := strings.ToLower(strings.Split(f.Tag.Get("json"), ",")[0])
		for _, bad := range countFieldNames {
			if name == bad || jsonTag == bad {
				t.Fatalf("%s 不应存在计数字段 %q（计数为服务端 owned，客户端不可注入）", typ.Name(), f.Name)
			}
		}
	}
}

// TestCount_ServerOwned_RequestTargetsHaveNoCountFields
// AC10 / ADR B3：请求体携带计数字段时应被忽略 —— 接口层 req 与应用层入参结构体均无计数字段。
func TestCount_ServerOwned_RequestTargetsHaveNoCountFields(t *testing.T) {
	// 1. 接口层 req：JSON 绑定目标不存在计数字段（伪造请求体无法注入）
	for _, target := range []any{
		types.CreateTaskReq{},
		types.UpdateTaskReq{},
		types.CreateProjectReq{},
		types.UpdateProjectReq{},
	} {
		assertNoCountField(t, reflect.TypeOf(target))
	}

	// 2. 伪造携带计数的请求体 → 真实绑定 + 真实转换器 → 应用层入参仍不含计数
	taskCreateBody := `{"name":"t","state":"todo","priority":"medium","checkItemCount":99,"commentCount":88,"subtaskCount":77}`
	var taskCreateReq types.CreateTaskReq
	if err := json.Unmarshal([]byte(taskCreateBody), &taskCreateReq); err != nil {
		t.Fatalf("绑定 CreateTaskReq: %v", err)
	}
	assertNoCountField(t, reflect.TypeOf(*toCreateTaskReq(&taskCreateReq)))

	taskUpdateBody := `{"name":"t2","checkItemCount":99,"commentCount":88,"subtaskCount":77}`
	var taskUpdateReq types.UpdateTaskReq
	if err := json.Unmarshal([]byte(taskUpdateBody), &taskUpdateReq); err != nil {
		t.Fatalf("绑定 UpdateTaskReq: %v", err)
	}
	assertNoCountField(t, reflect.TypeOf(*toUpdateTaskReq(&taskUpdateReq)))

	projectBody := `{"name":"p","taskCount":99}`
	var projectCreateReq types.CreateProjectReq
	if err := json.Unmarshal([]byte(projectBody), &projectCreateReq); err != nil {
		t.Fatalf("绑定 CreateProjectReq: %v", err)
	}
	assertNoCountField(t, reflect.TypeOf(*toCreateProjectInput(&projectCreateReq)))
	assertNoCountField(t, reflect.TypeOf(taskDto.CreateTaskReq{}))
	assertNoCountField(t, reflect.TypeOf(taskDto.UpdateTaskReq{}))
	assertNoCountField(t, reflect.TypeOf(projectDto.CreateProjectReq{}))
	assertNoCountField(t, reflect.TypeOf(projectDto.UpdateProjectReq{}))
}
