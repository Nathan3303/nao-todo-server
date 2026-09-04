package task

import (
	"reflect"
	"testing"
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
