package valueobjects

import (
	"strings"
	"testing"
)

func TestCreateTag_Validate(t *testing.T) {
	longName := strings.Repeat("a", 65)
	longDesc := strings.Repeat("b", 513)
	longColor := strings.Repeat("c", 17)

	tests := []struct {
		name    string
		vo      *CreateTag
		wantErr string
	}{
		{
			name: "valid tag",
			vo:   &CreateTag{Name: "urgent", Description: "Urgent tasks", Color: "#ff0000"},
		},
		{
			name:    "empty name",
			vo:      &CreateTag{Name: "", Color: "#ff0000"},
			wantErr: "标签名称不能为空",
		},
		{
			name:    "name too long",
			vo:      &CreateTag{Name: longName, Color: "#ff0000"},
			wantErr: "标签名称长度不能超过64个字符",
		},
		{
			name:    "description too long",
			vo:      &CreateTag{Name: "tag", Description: longDesc, Color: "#ff0000"},
			wantErr: "标签描述长度不能超过512个字符",
		},
		{
			name:    "empty color",
			vo:      &CreateTag{Name: "tag", Color: ""},
			wantErr: "标签颜色不能为空",
		},
		{
			name:    "color too long",
			vo:      &CreateTag{Name: "tag", Color: longColor},
			wantErr: "标签颜色长度不能超过16个字符",
		},
		{
			name: "max length name",
			vo:   &CreateTag{Name: strings.Repeat("a", 64), Color: "#ff0000"},
		},
		{
			name: "max length description",
			vo:   &CreateTag{Name: "tag", Description: strings.Repeat("b", 512), Color: "#ff0000"},
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
