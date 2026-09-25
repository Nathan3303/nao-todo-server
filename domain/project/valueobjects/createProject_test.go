package valueobjects

import (
	"strings"
	"testing"
)

func TestCreateProject_Validate(t *testing.T) {
	longName := strings.Repeat("a", 129)
	longDesc := strings.Repeat("b", 513)

	tests := []struct {
		name    string
		vo      *CreateProject
		wantErr string
	}{
		{
			name: "valid project",
			vo:   &CreateProject{UserId: 1, Name: "Valid Project", Description: "A project"},
		},
		{
			name:    "zero user id",
			vo:      &CreateProject{UserId: 0, Name: "Project"},
			wantErr: "用户 ID 不能为空",
		},
		{
			name:    "empty name",
			vo:      &CreateProject{UserId: 1, Name: ""},
			wantErr: "项目名称不能为空",
		},
		{
			name:    "name too long",
			vo:      &CreateProject{UserId: 1, Name: longName},
			wantErr: "项目名称不能超过128个字符",
		},
		{
			name:    "description too long",
			vo:      &CreateProject{UserId: 1, Name: "Valid", Description: longDesc},
			wantErr: "项目描述不能超过512个字符",
		},
		{
			name: "optional description empty",
			vo:   &CreateProject{UserId: 1, Name: "Valid", Description: ""},
		},
		{
			name: "max length name",
			vo:   &CreateProject{UserId: 1, Name: strings.Repeat("a", 128)},
		},
		{
			name: "max length description",
			vo:   &CreateProject{UserId: 1, Name: "Valid", Description: strings.Repeat("c", 512)},
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
