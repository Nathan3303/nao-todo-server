package task

import (
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
