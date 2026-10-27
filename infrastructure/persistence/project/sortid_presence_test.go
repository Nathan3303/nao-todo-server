package project

import (
	"testing"

	"naotodoserver/domain/project/valueobjects"
)

// TestCreateProjectVOToUpdateMapSortIdPresence F-06：
//   - 未提供（SortIdSet=false）⇒ update map 不出现 SortId（保留库中原值，修复每次 push 归零）
//   - 显式 0 / 非 0 ⇒ 照常写入（0 是合法排序值）
func TestCreateProjectVOToUpdateMapSortIdPresence(t *testing.T) {
	cases := []struct {
		name    string
		set     bool
		sortId  uint16
		wantKey bool
	}{
		{"未提供", false, 300, false},
		{"显式 0", true, 0, true},
		{"显式非 0", true, 700, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vo := &valueobjects.CreateProject{
				UserId: 1001, Name: "清单", SortId: tc.sortId, SortIdSet: tc.set,
			}
			updateMap := CreateProjectVOToUpdateMap(vo)
			got, ok := updateMap["SortId"]
			if ok != tc.wantKey {
				t.Fatalf("updateMap 含 SortId = %v, want %v（map=%+v）", ok, tc.wantKey, updateMap)
			}
			if !tc.wantKey {
				return
			}
			if got != tc.sortId {
				t.Fatalf("SortId = %v, want %d", got, tc.sortId)
			}
		})
	}
}
