// T466 单测（持久层）：CreateProjectVOToUpdateMap / CreateProjectValueObject2Model 的
// deactived_at 三态落点（DEF-46）。
//
// 「缺省 ⇒ 不产键 ⇒ 不写列」是关键回归锁：REST create 与旧客户端推送都不带该值，
// 一旦键恒在，会把服务端停用态抹成 NULL（并与 UpdateState 的 Delete/Restore 语义冲突）。
package project

import (
	"database/sql"
	"testing"
	"time"

	"naotodoserver/domain/project/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

func newCreateProjectVOWithDeactived(deactived domaintypes.NullableTime) *valueobjects.CreateProject {
	return &valueobjects.CreateProject{
		Id:          8001,
		UserId:      1001,
		Name:        "清单",
		Description: "",
		DeactivedAt: deactived,
	}
}

// TestCreateProjectVOToUpdateMapDeactivedAtTriState 三态：缺省不产键；显式清空写 NULL；值写时间。
func TestCreateProjectVOToUpdateMapDeactivedAtTriState(t *testing.T) {
	cases := []struct {
		name      string
		deactived domaintypes.NullableTime
		wantKey   bool
		wantNull  bool
	}{
		{"absent", domaintypes.NewNullableTimeNull(), false, false},
		{"empty-struct", domaintypes.NullableTime{}, false, false},
		{"set-null", domaintypes.NewNullableTimeSetToNull(), true, true},
		{"value", domaintypes.NewNullableTimeByTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			updateMap := CreateProjectVOToUpdateMap(newCreateProjectVOWithDeactived(tc.deactived))
			got, ok := updateMap["deactived_at"]
			if ok != tc.wantKey {
				t.Fatalf("updateMap 含 deactived_at = %v, want %v（map=%+v）", ok, tc.wantKey, updateMap)
			}
			if !tc.wantKey {
				return
			}
			nt, isNullTime := got.(sql.NullTime)
			if !isNullTime {
				t.Fatalf("deactived_at 值类型 = %T, want sql.NullTime", got)
			}
			if nt.Valid == tc.wantNull {
				t.Fatalf("sql.NullTime.Valid = %v, want %v（Valid=false ⇒ 写 NULL）", nt.Valid, !tc.wantNull)
			}
		})
	}
}

// TestCreateProjectValueObject2ModelDeactivedAt 新建（insert）路径：三态同样落到模型列。
func TestCreateProjectValueObject2ModelDeactivedAt(t *testing.T) {
	absent := CreateProjectValueObject2Model(newCreateProjectVOWithDeactived(domaintypes.NewNullableTimeNull()))
	if absent.DeactivedAt.Valid {
		t.Fatalf("缺省不应写 deactived_at: %+v", absent.DeactivedAt)
	}
	at := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	got := CreateProjectValueObject2Model(newCreateProjectVOWithDeactived(domaintypes.NewNullableTimeByTime(at)))
	if !got.DeactivedAt.Valid || !got.DeactivedAt.Time.Equal(at) {
		t.Fatalf("deactived_at = %+v, want %v", got.DeactivedAt, at)
	}
}
