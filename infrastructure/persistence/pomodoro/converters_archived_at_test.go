// T466 单测（持久层）：CreatePomodoroVOToUpdateMap / CreatePomodoroVOToModel 的
// archived_at 三态落点（DEF-43）。
//
// 「缺省 ⇒ 不产键 ⇒ 不写列」是关键回归锁：REST create 与旧客户端推送都不带该值，
// 一旦键恒在，会把服务端归档态抹成 NULL。
package pomodoro

import (
	"database/sql"
	"testing"
	"time"

	"naotodoserver/domain/types"
)

// TestCreatePomodoroVOToUpdateMap_ArchivedAt 三态：缺省不产键；显式清空写 NULL；值写时间。
func TestCreatePomodoroVOToUpdateMap_ArchivedAt(t *testing.T) {
	cases := []struct {
		name       string
		archivedAt types.NullableTime
		wantKey    bool
		wantNull   bool
	}{
		{"absent", types.NewNullableTimeNull(), false, false},
		{"empty-struct", types.NullableTime{}, false, false},
		{"set-null", types.NewNullableTimeSetToNull(), true, true},
		{"value", types.NewNullableTimeByTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)), true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vo := newPomodoroVO()
			vo.ArchivedAt = tc.archivedAt
			updateMap := CreatePomodoroVOToUpdateMap(vo)
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

// TestCreatePomodoroVOToModel_ArchivedAt 新建（insert）路径：三态同样落到模型列。
func TestCreatePomodoroVOToModel_ArchivedAt(t *testing.T) {
	vo := newPomodoroVO()
	vo.ArchivedAt = types.NewNullableTimeNull()
	if m := CreatePomodoroVOToModel(vo); m.ArchivedAt.Valid {
		t.Fatalf("缺省不应写 archived_at: %+v", m.ArchivedAt)
	}
	at := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	vo.ArchivedAt = types.NewNullableTimeByTime(at)
	m := CreatePomodoroVOToModel(vo)
	if !m.ArchivedAt.Valid || !m.ArchivedAt.Time.Equal(at) {
		t.Fatalf("archived_at = %+v, want %v", m.ArchivedAt, at)
	}
}
