package pomodoro

import (
	"testing"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/domain/types"
)

func newPomodoroVO() *valueobjects.CreatePomodoro {
	return &valueobjects.CreatePomodoro{
		Id:   1,
		Type: entities.PomodoroTypeFocus,
		Name: "专注",
	}
}

// TestCreatePomodoroVOToModel_DeletedAt 创建模型携带/未携带 DeletedAt
func TestCreatePomodoroVOToModel_DeletedAt(t *testing.T) {
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
			vo := newPomodoroVO()
			vo.DeletedAt = tt.deletedAt
			m := CreatePomodoroVOToModel(vo)
			if got := m.DeletedAt.Valid; got != tt.wantValid {
				t.Errorf("DeletedAt.Valid = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

// TestCreatePomodoroVOToUpdateMap_DeletedAt 更新映射仅携带删除时间时写入 deleted_at
func TestCreatePomodoroVOToUpdateMap_DeletedAt(t *testing.T) {
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
			vo := newPomodoroVO()
			vo.DeletedAt = tt.deletedAt
			m := CreatePomodoroVOToUpdateMap(vo)
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
