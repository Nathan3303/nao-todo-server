package entities

import (
	"testing"
	"time"

	"naotodoserver/domain/types"
)

func TestProject_Archive(t *testing.T) {
	tests := []struct {
		name       string
		archivedAt types.NullableTime
	}{
		{name: "未归档", archivedAt: types.NewNullableTimeNull()},
		{name: "已归档 重复归档幂等", archivedAt: types.NewNullableTimeByTime(time.Now().Add(-time.Hour))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &Project{ArchivedAt: tt.archivedAt}
			project.Archive()
			if _, ok := project.ArchivedAt.Value(); !ok {
				t.Errorf("Archive() 后 ArchivedAt 无效，want 有效且非空")
			}
			if project.ArchivedAt.IsNull {
				t.Errorf("Archive() 后 ArchivedAt.IsNull = true, want false")
			}
		})
	}
}

func TestProject_Unarchive(t *testing.T) {
	tests := []struct {
		name       string
		archivedAt types.NullableTime
	}{
		{name: "已归档", archivedAt: types.NewNullableTimeByTime(time.Now())},
		{name: "未归档 幂等", archivedAt: types.NewNullableTimeNull()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &Project{ArchivedAt: tt.archivedAt}
			project.Unarchive()
			if _, ok := project.ArchivedAt.Value(); ok {
				t.Errorf("Unarchive() 后 ArchivedAt 有效，want 为空")
			}
			if !project.ArchivedAt.IsNull {
				t.Errorf("Unarchive() 后 ArchivedAt.IsNull = false, want true")
			}
		})
	}
}

func TestProject_Delete(t *testing.T) {
	tests := []struct {
		name        string
		deactivedAt types.NullableTime
	}{
		{name: "未删除", deactivedAt: types.NewNullableTimeNull()},
		{name: "已删除 重复删除幂等", deactivedAt: types.NewNullableTimeByTime(time.Now().Add(-time.Hour))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &Project{DeactivedAt: tt.deactivedAt}
			project.Delete()
			if _, ok := project.DeactivedAt.Value(); !ok {
				t.Errorf("Delete() 后 DeactivedAt 无效，want 有效且非空")
			}
			if project.DeactivedAt.IsNull {
				t.Errorf("Delete() 后 DeactivedAt.IsNull = true, want false")
			}
		})
	}
}

func TestProject_Restore(t *testing.T) {
	tests := []struct {
		name        string
		deactivedAt types.NullableTime
	}{
		{name: "已删除", deactivedAt: types.NewNullableTimeByTime(time.Now())},
		{name: "未删除 幂等", deactivedAt: types.NewNullableTimeNull()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &Project{DeactivedAt: tt.deactivedAt}
			project.Restore()
			if _, ok := project.DeactivedAt.Value(); ok {
				t.Errorf("Restore() 后 DeactivedAt 有效，want 为空")
			}
			if !project.DeactivedAt.IsNull {
				t.Errorf("Restore() 后 DeactivedAt.IsNull = false, want true")
			}
		})
	}
}
