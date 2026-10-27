package task

import (
	"testing"

	"naotodoserver/domain/task/valueobjects"
)

// TestTaskCheckItemVOToUpdateMapDescriptionPresence F-06 核心回归锁：
//   - 未提供（DescriptionSet=false）⇒ update map 不出现 Description（保留库中原值）
//   - 显式 "" ⇒ 写入空串（用户能清空备注）
//   - 值 ⇒ 写入该值
func TestTaskCheckItemVOToUpdateMapDescriptionPresence(t *testing.T) {
	cases := []struct {
		name    string
		set     bool
		desc    string
		wantKey bool
		wantVal string
	}{
		{"未提供", false, "", false, ""},
		{"显式置空", true, "", true, ""},
		{"显式值", true, "备注B", true, "备注B"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vo := &valueobjects.CreateTaskCheckItem{
				TaskId: 1, Name: "检查项", Description: tc.desc, DescriptionSet: tc.set,
			}
			updateMap := TaskCheckItemVOToUpdateMap(vo)
			got, ok := updateMap["Description"]
			if ok != tc.wantKey {
				t.Fatalf("updateMap 含 Description = %v, want %v（map=%+v）", ok, tc.wantKey, updateMap)
			}
			if !tc.wantKey {
				return
			}
			if got != tc.wantVal {
				t.Fatalf("Description = %v, want %q", got, tc.wantVal)
			}
		})
	}
}

// TestTaskCommentVOToUpdateMapPresence F-06：attachments / isTopUp 未提供 ⇒ 不产键；
// 显式提供（含空数组 / false）⇒ 照常写入。
func TestTaskCommentVOToUpdateMapPresence(t *testing.T) {
	t.Run("未提供", func(t *testing.T) {
		updateMap := TaskCommentVOToUpdateMap(&valueobjects.CreateTaskComment{Content: "c"})
		if _, ok := updateMap["Attachments"]; ok {
			t.Fatalf("未提供 attachments 不应产键: %+v", updateMap)
		}
		if _, ok := updateMap["IsTopUp"]; ok {
			t.Fatalf("未提供 isTopUp 不应产键: %+v", updateMap)
		}
	})

	t.Run("显式提供", func(t *testing.T) {
		updateMap := TaskCommentVOToUpdateMap(&valueobjects.CreateTaskComment{
			Content:        "c",
			Attachments:    []string{"f1"},
			AttachmentsSet: true,
			IsTopUp:        true,
			IsTopUpSet:     true,
		})
		if got := updateMap["Attachments"]; got != `["f1"]` {
			t.Fatalf("Attachments = %v, want [\"f1\"]", got)
		}
		if got := updateMap["IsTopUp"]; got != true {
			t.Fatalf("IsTopUp = %v, want true", got)
		}
	})

	t.Run("显式清空", func(t *testing.T) {
		updateMap := TaskCommentVOToUpdateMap(&valueobjects.CreateTaskComment{
			Content:        "c",
			Attachments:    []string{},
			AttachmentsSet: true,
			IsTopUp:        false,
			IsTopUpSet:     true,
		})
		if got := updateMap["Attachments"]; got != "[]" {
			t.Fatalf("Attachments = %v, want []", got)
		}
		if got := updateMap["IsTopUp"]; got != false {
			t.Fatalf("IsTopUp = %v, want false", got)
		}
	})
}
