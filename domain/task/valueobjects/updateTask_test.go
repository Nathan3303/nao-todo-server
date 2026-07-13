package valueobjects

import (
	"strings"
	"testing"

	"naotodoserver/domain/task/entities"
)

func strPtr(s string) *string       { return &s }
func uint8Ptr(v uint8) *uint8       { return &v }
func uint16Ptr(v uint16) *uint16    { return &v }
func taskStatePtr(v entities.TaskState) *entities.TaskState     { return &v }
func taskPriorityPtr(v entities.TaskPriority) *entities.TaskPriority { return &v }

func TestUpdateTask_Validate(t *testing.T) {
	longName := strings.Repeat("a", 257)
	longDesc := strings.Repeat("b", 513)

	tests := []struct {
		name    string
		vo      *UpdateTask
		wantErr string
	}{
		{
			name: "all nil (no update fields)",
			vo:   &UpdateTask{},
		},
		{
			name: "valid name update",
			vo:   &UpdateTask{Name: strPtr("Updated Name")},
		},
		{
			name:    "empty name",
			vo:      &UpdateTask{Name: strPtr("")},
			wantErr: "任务名称不能为空",
		},
		{
			name:    "name too long",
			vo:      &UpdateTask{Name: strPtr(longName)},
			wantErr: "任务名称最多256个字符",
		},
		{
			name: "valid description update",
			vo:   &UpdateTask{Description: strPtr("Updated desc")},
		},
		{
			name:    "description too long",
			vo:      &UpdateTask{Description: strPtr(longDesc)},
			wantErr: "任务描述最多512个字符",
		},
		{
			name: "valid state update",
			vo:   &UpdateTask{State: taskStatePtr(entities.TaskStateCompleted)},
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

func TestUpdateTask_Trim(t *testing.T) {
	tests := []struct {
		name      string
		inputName *string
		inputDesc *string
		wantName  *string
		wantDesc  *string
	}{
		{
			name:      "trim name and description",
			inputName: strPtr("  hello  "),
			inputDesc: strPtr("  world  "),
			wantName:  strPtr("hello"),
			wantDesc:  strPtr("world"),
		},
		{
			name:      "nil fields unchanged",
			inputName: nil,
			inputDesc: nil,
			wantName:  nil,
			wantDesc:  nil,
		},
		{
			name:      "only name set",
			inputName: strPtr("  spaced  "),
			inputDesc: nil,
			wantName:  strPtr("spaced"),
			wantDesc:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo := &UpdateTask{
				Name:        tt.inputName,
				Description: tt.inputDesc,
			}
			vo.Trim()

			if tt.wantName == nil && vo.Name != nil {
				t.Errorf("Trim() Name = %v, want nil", *vo.Name)
			}
			if tt.wantName != nil && (vo.Name == nil || *vo.Name != *tt.wantName) {
				got := ""
				if vo.Name != nil {
					got = *vo.Name
				}
				t.Errorf("Trim() Name = %q, want %q", got, *tt.wantName)
			}
			if tt.wantDesc == nil && vo.Description != nil {
				t.Errorf("Trim() Description = %v, want nil", *vo.Description)
			}
			if tt.wantDesc != nil && (vo.Description == nil || *vo.Description != *tt.wantDesc) {
				got := ""
				if vo.Description != nil {
					got = *vo.Description
				}
				t.Errorf("Trim() Description = %q, want %q", got, *tt.wantDesc)
			}
		})
	}
}
