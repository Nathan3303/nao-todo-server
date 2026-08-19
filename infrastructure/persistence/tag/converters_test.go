package tag

import (
	"testing"
	"time"

	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

func TestTagModel2Entity_DeletedAt(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name       string
		deletedAt  gorm.DeletedAt
		wantString string
	}{
		{name: "已软删 带删除时间", deletedAt: gorm.DeletedAt{Time: now, Valid: true}, wantString: "2026-08-13T10:30:00Z"},
		{name: "未删除", deletedAt: gorm.DeletedAt{}, wantString: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &models.Tag{
				ModelBase: models.ModelBase{
					ID:        1,
					DeletedAt: tt.deletedAt,
				},
				Name: "测试标签",
			}
			e := TagModel2Entity(m)
			got := e.DeletedAt.ToString(time.RFC3339)
			if got != tt.wantString {
				t.Errorf("TagModel2Entity DeletedAt.ToString() = %q, want %q", got, tt.wantString)
			}
		})
	}
}

func TestTagPreferenceModel2Entity_DeletedAt(t *testing.T) {
	now := time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		deletedAt  gorm.DeletedAt
		wantString string
	}{
		{name: "已软删 带删除时间", deletedAt: gorm.DeletedAt{Time: now, Valid: true}, wantString: "2026-08-13T11:00:00Z"},
		{name: "未删除", deletedAt: gorm.DeletedAt{}, wantString: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &models.TagPreference{
				ModelBase: models.ModelBase{
					ID:        1,
					DeletedAt: tt.deletedAt,
				},
				TagId: 1,
			}
			e := TagPreferenceModel2Entity(m)
			got := e.DeletedAt.ToString(time.RFC3339)
			if got != tt.wantString {
				t.Errorf("TagPreferenceModel2Entity DeletedAt.ToString() = %q, want %q", got, tt.wantString)
			}
		})
	}
}
