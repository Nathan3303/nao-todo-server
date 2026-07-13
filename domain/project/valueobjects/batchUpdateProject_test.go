package valueobjects

import (
	"strings"
	"testing"
)

func TestBatchUpdateProject_Validate(t *testing.T) {
	longName := strings.Repeat("a", 129)
	longDesc := strings.Repeat("b", 513)

	tests := []struct {
		name    string
		vo      *BatchUpdateProject
		wantErr string
	}{
		{
			name: "valid update",
			vo:   &BatchUpdateProject{Id: 1, Name: strPtr("Updated"), Description: strPtr("Desc")},
		},
		{
			name:    "invalid id (zero)",
			vo:      &BatchUpdateProject{Id: 0},
			wantErr: "项目 ID 无效",
		},
		{
			name:    "invalid id (negative)",
			vo:      &BatchUpdateProject{Id: -1},
			wantErr: "项目 ID 无效",
		},
		{
			name:    "name too long",
			vo:      &BatchUpdateProject{Id: 1, Name: strPtr(longName)},
			wantErr: "项目名称不能超过128个字符",
		},
		{
			name:    "description too long",
			vo:      &BatchUpdateProject{Id: 1, Description: strPtr(longDesc)},
			wantErr: "项目描述不能超过512个字符",
		},
		{
			name:    "sort id is 0",
			vo:      &BatchUpdateProject{Id: 1, SortId: uint16Ptr(0)},
			wantErr: "排序ID不能为0",
		},
		{
			name: "sort id valid",
			vo:   &BatchUpdateProject{Id: 1, SortId: uint16Ptr(100)},
		},
		{
			name: "all optional fields nil",
			vo:   &BatchUpdateProject{Id: 1},
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
