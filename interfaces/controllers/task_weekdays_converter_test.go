// T206 边界回归：toUint8Weekdays 的越界输入必须按 domain 既有规则忽略（entities.WeekdaysToBitmask 的
// ">6 忽略"），不得因 int→uint8 截断被隐式换成合法星期（原 256→周日、257→周一、-256→周日），
// 同时保证 0..6 与 nil / 空输入的行为逐字不变。
package controllers

import (
	"reflect"
	"testing"
)

func TestToUint8Weekdays_Boundary(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []uint8
	}{
		{"nil 保持 nil", nil, nil},
		{"空切片返回非 nil 空切片", []int{}, []uint8{}},
		{"有效域 0..6 原样透传", []int{0, 1, 2, 3, 4, 5, 6}, []uint8{0, 1, 2, 3, 4, 5, 6}},
		{"单个有效值", []int{6}, []uint8{6}},
		{"越界值忽略（原本就被 domain 忽略）", []int{7, 255, 263, 300}, []uint8{}},
		{"截断会被误当合法星期的值忽略", []int{256, 257, 262}, []uint8{}},
		{"负数忽略", []int{-1, -256, -257}, []uint8{}},
		{"混合：只保留 0..6", []int{256, 1, -256, 3, 7, 6}, []uint8{1, 3, 6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toUint8Weekdays(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("toUint8Weekdays(%v) = %v, want %v", tt.in, got, tt.want)
			}
			// nil / 空切片语义：nil 进 nil 出，空切片进非 nil 空切片出（update 路径靠 != nil 判定）
			if tt.in == nil && got != nil {
				t.Fatalf("nil 输入应返回 nil，实际 %v（非 nil）", got)
			}
			if tt.in != nil && len(tt.in) == 0 && got == nil {
				t.Fatal("空切片输入应返回非 nil 空切片，实际 nil")
			}
		})
	}
}
