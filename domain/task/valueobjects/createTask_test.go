package valueobjects

import (
	"strings"
	"testing"
)

func TestCreateTask_Validate(t *testing.T) {
	longName := strings.Repeat("a", 257)
	longDesc := strings.Repeat("b", 513)

	tests := []struct {
		name    string
		vo      *CreateTask
		wantErr string
	}{
		{
			name: "valid task",
			vo:   &CreateTask{Name: "Valid Task", Description: "A description"},
		},
		{
			name:    "empty name",
			vo:      &CreateTask{Name: "", Description: "desc"},
			wantErr: "任务名称不能为空",
		},
		{
			name:    "name too long",
			vo:      &CreateTask{Name: longName},
			wantErr: "任务名称最多256个字符",
		},
		{
			name:    "description too long",
			vo:      &CreateTask{Name: "Valid", Description: longDesc},
			wantErr: "任务描述最多512个字符",
		},
		{
			name: "optional description empty",
			vo:   &CreateTask{Name: "Valid", Description: ""},
		},
		{
			name: "max length name",
			vo:   &CreateTask{Name: strings.Repeat("a", 256)},
		},
		{
			name: "max length description",
			vo:   &CreateTask{Name: "Valid", Description: strings.Repeat("c", 512)},
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
