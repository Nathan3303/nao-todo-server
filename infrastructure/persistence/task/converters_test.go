package task

import (
	"database/sql"
	"testing"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/domain/types"
)

func newConvertTaskVO() *valueobjects.CreateTask {
	return &valueobjects.CreateTask{
		Id:       1,
		Name:     "任务",
		State:    entities.TaskStatePending,
		Priority: entities.TaskPriorityMedium,
	}
}

// TestCreateTaskValueObjectToModel_DeletedAt 创建模型携带/未携带 DeletedAt
func TestCreateTaskValueObjectToModel_DeletedAt(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		deletedAt types.NullableTime
		wantValid bool
	}{
		{name: "携带删除时间", deletedAt: types.NewNullableTimeByTime(now), wantValid: true},
		{name: "未携带", deletedAt: types.NullableTime{}, wantValid: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo := newConvertTaskVO()
			vo.DeletedAt = tt.deletedAt
			m := CreateTaskValueObjectToModel(1, vo)
			if got := m.DeletedAt.Valid; got != tt.wantValid {
				t.Errorf("DeletedAt.Valid = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestCreateTaskVOToUpdateMap_DeletedAt 更新映射仅携带删除时间时写入 deleted_at
func TestCreateTaskVOToUpdateMap_DeletedAt(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		deletedAt   types.NullableTime
		wantWritten bool
	}{
		{name: "携带删除时间", deletedAt: types.NewNullableTimeByTime(now), wantWritten: true},
		{name: "未携带", deletedAt: types.NullableTime{}, wantWritten: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo := newConvertTaskVO()
			vo.DeletedAt = tt.deletedAt
			m := CreateTaskVOToUpdateMap(vo)
			_, ok := m["deleted_at"]
			if ok != tt.wantWritten {
				t.Errorf("updateMap 含 deleted_at = %v, want %v", ok, tt.wantWritten)
			}
			if tt.wantWritten && m["deleted_at"] == nil {
				t.Error("deleted_at 值不应为 nil")
			}
		})
	}
}

// TestCreateTaskVOToUpdateMap_NullableTimeSemantics SYNC-DEF-01：可空时间字段
// 缺省（Valid=false）不写列；显式空串（Valid=true,IsNull=true）写 NULL；有效值写时间。
func TestCreateTaskVOToUpdateMap_NullableTimeSemantics(t *testing.T) {
	t.Run("缺省不写列", func(t *testing.T) {
		vo := newConvertTaskVO()
		m := CreateTaskVOToUpdateMap(vo)
		for _, k := range []string{"StartAt", "EndAt", "RemindAt"} {
			if _, ok := m[k]; ok {
				t.Fatalf("缺省的 %s 不应进入 updateMap", k)
			}
		}
	})

	t.Run("显式空串写 NULL", func(t *testing.T) {
		vo := newConvertTaskVO()
		vo.StartAt = types.NewNullableTimeByTimeStr("")
		vo.EndAt = types.NewNullableTimeByTimeStr("")
		vo.RemindAt = types.NewNullableTimeByTimeStr("")
		m := CreateTaskVOToUpdateMap(vo)
		for _, k := range []string{"StartAt", "EndAt"} {
			st, ok := m[k].(sql.NullTime)
			if !ok {
				t.Fatalf("%s 应进入 updateMap 且为 sql.NullTime, got %#v", k, m[k])
			}
			if st.Valid {
				t.Fatalf("%s 显式空串应写 NULL, got %v", k, st.Time)
			}
		}
		if rt, ok := m["RemindAt"].(sql.NullTime); !ok || rt.Valid {
			t.Fatalf("RemindAt 显式空串应写 NULL, got %#v", m["RemindAt"])
		}
	})
}

// --- DEF-SYNC-04：创建/覆盖写路径秒级往返不变（往返归一） ---

// TestCreateTaskValueObjectToModel_RoundtripSecondLevel 提供 startAt+endAt（含毫秒）
// ⇒ 落库模型与实体往返后秒级瞬时不变（毫秒截断而非四舍五入，不产生下一秒漂移）
func TestCreateTaskValueObjectToModel_RoundtripSecondLevel(t *testing.T) {
	// .900 的毫秒：若 MySQL 四舍五入会进到下一秒（41→42），显式截断必须保持 41
	startAt, err := time.Parse(time.RFC3339Nano, "2026-09-11T16:28:41.900+08:00")
	if err != nil {
		t.Fatal(err)
	}
	endAt, err := time.Parse(time.RFC3339Nano, "2026-09-11T20:00:00.123+08:00")
	if err != nil {
		t.Fatal(err)
	}
	vo := newConvertTaskVO()
	vo.StartAt = types.NewNullableTimeByTime(startAt)
	vo.EndAt = types.NewNullableTimeByTime(endAt)

	m := CreateTaskValueObjectToModel(1, vo)
	if !m.StartAt.Valid || !m.EndAt.Valid {
		t.Fatal("startAt/endAt 应有效")
	}
	// 落库值：秒级瞬时与客户端一致，且不含小数秒
	if got := m.StartAt.Time.Truncate(time.Second).Unix(); got != startAt.Unix() {
		t.Fatalf("落库 startAt 秒级漂移: got %d, want %d", got, startAt.Unix())
	}
	if got := m.EndAt.Time.Truncate(time.Second).Unix(); got != endAt.Unix() {
		t.Fatalf("落库 endAt 秒级漂移: got %d, want %d", got, endAt.Unix())
	}
	if m.StartAt.Time.Nanosecond() != 0 {
		t.Fatalf("startAt 应截断到秒（毫秒截断不算破坏），got 纳秒 %d", m.StartAt.Time.Nanosecond())
	}

	// 返回路径：模型 → 实体（等价于 create 响应 / pull 拉取）秒级瞬时不变
	e := TaskModel2Entity(m)
	gotStart, ok := e.StartAt.Value()
	if !ok {
		t.Fatal("实体 startAt 应有效")
	}
	if gotStart.Unix() != startAt.Unix() {
		t.Fatalf("往返后 startAt 秒级漂移: got %d, want %d", gotStart.Unix(), startAt.Unix())
	}
}

// TestCreateTaskVOToUpdateMap_StartEndSecondLevel 覆盖分支（repoImpl 覆盖写）同样秒级截断
func TestCreateTaskVOToUpdateMap_StartEndSecondLevel(t *testing.T) {
	startAt, err := time.Parse(time.RFC3339Nano, "2026-09-11T16:28:41.900+08:00")
	if err != nil {
		t.Fatal(err)
	}
	vo := newConvertTaskVO()
	vo.StartAt = types.NewNullableTimeByTime(startAt)

	m := CreateTaskVOToUpdateMap(vo)
	st := m["StartAt"].(sql.NullTime)
	if !st.Valid {
		t.Fatal("updateMap StartAt 应有效")
	}
	if st.Time.Truncate(time.Second).Unix() != startAt.Unix() {
		t.Fatalf("覆盖分支 startAt 秒级漂移: got %d, want %d", st.Time.Truncate(time.Second).Unix(), startAt.Unix())
	}
	if st.Time.Nanosecond() != 0 {
		t.Fatalf("覆盖分支 startAt 应截断到秒，got 纳秒 %d", st.Time.Nanosecond())
	}
}

// TestCreateTaskVOToUpdateMap_StatusTimestamps DEF-SYNC-06：Create VO 的状态时间戳
// （ArchivedAt/StarMarkAt/GivenUpAt）按三态写入 upsert 覆盖映射：
// 缺省不写列；显式空串写 NULL；有效值写截断到秒的时间。
func TestCreateTaskVOToUpdateMap_StatusTimestamps(t *testing.T) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 900*int(time.Millisecond), time.UTC)
	cols := []struct {
		name string
		set  func(*valueobjects.CreateTask, types.NullableTime)
		key  string
	}{
		{"ArchivedAt", func(vo *valueobjects.CreateTask, nt types.NullableTime) { vo.ArchivedAt = nt }, "ArchivedAt"},
		{"StarMarkAt", func(vo *valueobjects.CreateTask, nt types.NullableTime) { vo.StarMarkAt = nt }, "StarMarkAt"},
		{"GivenUpAt", func(vo *valueobjects.CreateTask, nt types.NullableTime) { vo.GivenUpAt = nt }, "GivenUpAt"},
	}
	for _, c := range cols {
		t.Run(c.name+"/缺省不写列", func(t *testing.T) {
			m := CreateTaskVOToUpdateMap(newConvertTaskVO())
			if _, ok := m[c.key]; ok {
				t.Fatalf("缺省的 %s 不应进入 updateMap", c.key)
			}
		})
		t.Run(c.name+"/空串写 NULL", func(t *testing.T) {
			vo := newConvertTaskVO()
			c.set(vo, types.NewNullableTimeByTimeStr(""))
			st, ok := CreateTaskVOToUpdateMap(vo)[c.key].(sql.NullTime)
			if !ok {
				t.Fatalf("%s 应进入 updateMap 且为 sql.NullTime", c.key)
			}
			if st.Valid {
				t.Fatalf("%s 显式空串应写 NULL, got %v", c.key, st.Time)
			}
		})
		t.Run(c.name+"/有效值写秒级", func(t *testing.T) {
			vo := newConvertTaskVO()
			c.set(vo, types.NewNullableTimeByTime(now))
			st, ok := CreateTaskVOToUpdateMap(vo)[c.key].(sql.NullTime)
			if !ok || !st.Valid {
				t.Fatalf("%s 有效值应写入有效 sql.NullTime, got %#v", c.key, CreateTaskVOToUpdateMap(vo)[c.key])
			}
			if st.Time.Nanosecond() != 0 || !st.Time.Equal(now.Truncate(time.Second)) {
				t.Fatalf("%s 应截断到秒, got %v", c.key, st.Time)
			}
		})
	}
}

// TestCreateTaskValueObjectToModel_StatusTimestamps 新建分支（CreateTaskValueObjectToModel）
// 同样须携带三状态时间戳，否则首次推送（服务端无既有行）仍会丢字段。
func TestCreateTaskValueObjectToModel_StatusTimestamps(t *testing.T) {
	now := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	vo := newConvertTaskVO()
	vo.ArchivedAt = types.NewNullableTimeByTime(now)
	vo.StarMarkAt = types.NewNullableTimeByTime(now.Add(time.Second))
	vo.GivenUpAt = types.NewNullableTimeByTime(now.Add(2 * time.Second))
	m := CreateTaskValueObjectToModel(1, vo)
	if !m.ArchivedAt.Valid || !m.StarMarkAt.Valid || !m.GivenUpAt.Valid {
		t.Fatalf("新建模型丢失状态时间戳: archived=%v star=%v givenUp=%v",
			m.ArchivedAt, m.StarMarkAt, m.GivenUpAt)
	}
	if !m.GivenUpAt.Time.Equal(now.Add(2 * time.Second)) {
		t.Fatalf("given_up_at 值错误: got %v", m.GivenUpAt.Time)
	}
}
