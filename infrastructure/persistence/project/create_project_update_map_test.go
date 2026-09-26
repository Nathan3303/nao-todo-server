// T322 单测（持久层）：CreateProjectVOToUpdateMap 的 archived_at 三态落点
//
// 「缺省 ⇒ 不产键 ⇒ 不写列」是本单最关键的回归锁：REST create 与旧客户端推送都不带该值，
// 一旦键恒在，会把服务端归档态抹成 NULL。
package project

import (
	"database/sql"
	"testing"
	"time"

	"naotodoserver/domain/project/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

func newCreateProjectVO(archived domaintypes.NullableTime) *valueobjects.CreateProject {
	return &valueobjects.CreateProject{
		Id:          8001,
		UserId:      1001,
		Name:        "清单",
		Description: "",
		ArchivedAt:  archived,
	}
}

// TestCreateProjectVOToUpdateMapArchivedAtTriState 三态：缺省不产键；显式清空写 NULL；值写时间。
func TestCreateProjectVOToUpdateMapArchivedAtTriState(t *testing.T) {
	cases := []struct {
		name     string
		archived domaintypes.NullableTime
		wantKey  bool
		wantNull bool
	}{
		{"absent", domaintypes.NewNullableTimeNull(), false, false},
		{"empty-struct", domaintypes.NullableTime{}, false, false},
		{"set-null", domaintypes.NewNullableTimeSetToNull(), true, true},
		{"value", domaintypes.NewNullableTimeByTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			updateMap := CreateProjectVOToUpdateMap(newCreateProjectVO(tc.archived))
			got, ok := updateMap["archived_at"]
			if ok != tc.wantKey {
				t.Fatalf("updateMap 含 archived_at = %v, want %v（map=%+v）", ok, tc.wantKey, updateMap)
			}
			if !tc.wantKey {
				return
			}
			nt, isNullTime := got.(sql.NullTime)
			if !isNullTime {
				t.Fatalf("archived_at 值类型 = %T, want sql.NullTime", got)
			}
			if nt.Valid == tc.wantNull {
				t.Fatalf("sql.NullTime.Valid = %v, want %v（Valid=false ⇒ 写 NULL）", nt.Valid, !tc.wantNull)
			}
		})
	}
}

// TestCreateProjectValueObject2ModelArchivedAt 新建（insert）路径：三态同样落到模型列。
func TestCreateProjectValueObject2ModelArchivedAt(t *testing.T) {
	absent := CreateProjectValueObject2Model(newCreateProjectVO(domaintypes.NewNullableTimeNull()))
	if absent.ArchivedAt.Valid {
		t.Fatalf("缺省不应写 archived_at: %+v", absent.ArchivedAt)
	}
	at := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	got := CreateProjectValueObject2Model(newCreateProjectVO(domaintypes.NewNullableTimeByTime(at)))
	if !got.ArchivedAt.Valid || !got.ArchivedAt.Time.Equal(at) {
		t.Fatalf("archived_at = %+v, want %v", got.ArchivedAt, at)
	}
}
