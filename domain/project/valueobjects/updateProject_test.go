package valueobjects

import (
	"strings"
	"testing"
)

func strPtr(s string) *string   { return &s }
func uint16Ptr(v uint16) *uint16 { return &v }

func TestUpdateProject_Validate(t *testing.T) {
	longName := strings.Repeat("a", 129)
	longDesc := strings.Repeat("b", 513)

	tests := []struct {
		name    string
		vo      *UpdateProject
		wantErr string
	}{
		{
			name: "all nil (no update fields)",
			vo:   &UpdateProject{},
		},
		{
			name: "valid name update",
			vo:   &UpdateProject{Name: strPtr("Updated Project")},
		},
		{
			name:    "empty name",
			vo:      &UpdateProject{Name: strPtr("")},
			wantErr: "项目名称不能为空",
		},
		{
			name:    "name too long",
			vo:      &UpdateProject{Name: strPtr(longName)},
			wantErr: "项目名称不能超过128个字符",
		},
		{
			name: "valid description update",
			vo:   &UpdateProject{Description: strPtr("Updated desc")},
		},
		{
			name:    "description too long",
			vo:      &UpdateProject{Description: strPtr(longDesc)},
			wantErr: "项目描述不能超过512个字符",
		},
		{
			name: "sort id zero (valid nil)",
			vo:   &UpdateProject{SortId: nil},
		},
		{
			name:    "sort id is 0",
			vo:      &UpdateProject{SortId: uint16Ptr(0)},
			wantErr: "排序ID不能为0",
		},
		{
			name: "sort id valid non-zero",
			vo:   &UpdateProject{SortId: uint16Ptr(100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vo.Validate()
			if tt.wantErr == "" && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
			if tt.wantErr != "" && err == nil {
				t.Errorf("Validate() want error %q, got nil", tt.wantErr)
			}
			if tt.wantErr != "" && err != nil && err.Error() != tt.wantErr {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
