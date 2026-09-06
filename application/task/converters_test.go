package task

import (
	"reflect"
	"testing"

	"naotodoserver/application/task/dto"
)

func TestSplitProjectIds(t *testing.T) {
	const userId int64 = 1001
	tests := []struct {
		name    string
		raw     string
		want    []int64
		wantErr bool
	}{
		{name: "空串 = 无过滤", raw: "", want: nil},
		{name: "单合法", raw: "111", want: []int64{111}},
		{name: "单 inbox", raw: "inbox", want: []int64{userId}},
		{name: "多值保留顺序", raw: "111,222,333", want: []int64{111, 222, 333}},
		{name: "inbox 混入多值", raw: "111,inbox,333", want: []int64{111, userId, 333}},
		{name: "非法段跳过", raw: "111,abc,333", want: []int64{111, 333}},
		{name: "全非法报错", raw: "abc,def", wantErr: true},
		{name: "纯逗号报错", raw: ",,", wantErr: true},
		{name: "去重", raw: "111,111,222", want: []int64{111, 222}},
		{name: "空白段与首尾空格", raw: " 111 , ,222", want: []int64{111, 222}},
		{name: "非正数段跳过", raw: "111,-5,222", want: []int64{111, 222}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitProjectIds(tt.raw, userId)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("splitProjectIds(%q) 期望报错，实际 nil", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitProjectIds(%q) 意外错误: %v", tt.raw, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitProjectIds(%q) = %v, 期望 %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestSplitTagIds(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "空串 = 无过滤", raw: "", want: nil},
		{name: "单值", raw: "tag1", want: []string{"tag1"}},
		{name: "多值保留顺序", raw: "tag1,tag2,tag3", want: []string{"tag1", "tag2", "tag3"}},
		{name: "空段去除", raw: "tag1,,tag2", want: []string{"tag1", "tag2"}},
		{name: "去重", raw: "tag1,tag1,tag2", want: []string{"tag1", "tag2"}},
		{name: "纯逗号 = 无过滤", raw: ",,,", want: nil},
		{name: "首尾空格", raw: " tag1 , tag2 ", want: []string{"tag1", "tag2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitTagIds(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitTagIds(%q) = %v, 期望 %v", tt.raw, got, tt.want)
			}
		})
	}
}

// TestUpdateTaskReqToValueObject_ProjectId 更新路径 projectId 归一语义（CAL-03 BE-1/2/3）：
// nil（缺省/null，反序列化后不可区分）= 不改；”/inbox = 默认收件箱 userId；合法 ID = 覆盖。
func TestUpdateTaskReqToValueObject_ProjectId(t *testing.T) {
	const userId int64 = 1001
	strPtr := func(s string) *string { return &s }
	int64Ptr := func(v int64) *int64 { return &v }
	tests := []struct {
		name      string
		projectId *string
		want      *int64
	}{
		{name: "缺省/nil = 不改", projectId: nil, want: nil},
		{name: "空串 = 收件箱 userId", projectId: strPtr(""), want: int64Ptr(userId)},
		{name: "inbox 字面量 = 收件箱 userId", projectId: strPtr("inbox"), want: int64Ptr(userId)},
		{name: "合法 ID = 覆盖", projectId: strPtr("222"), want: int64Ptr(222)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo, err := UpdateTaskReqToValueObject(userId, &dto.UpdateTaskReq{ProjectId: tt.projectId})
			if err != nil {
				t.Fatalf("UpdateTaskReqToValueObject 意外错误: %v", err)
			}
			if !reflect.DeepEqual(vo.ProjectId, tt.want) {
				t.Fatalf("ProjectId 转换 = %v, 期望 %v", vo.ProjectId, tt.want)
			}
		})
	}
}
