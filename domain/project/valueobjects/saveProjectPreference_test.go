package valueobjects

import "testing"

func TestSaveProjectPreference_Validate(t *testing.T) {
	tests := []struct {
		name    string
		vo      *SaveProjectPreference
		wantErr string
	}{
		{
			name: "valid table",
			vo:   &SaveProjectPreference{ViewType: "table", GetOptions: "all", Columns: "name,desc"},
		},
		{
			name: "valid list",
			vo:   &SaveProjectPreference{ViewType: "list", GetOptions: "all", Columns: "name"},
		},
		{
			name: "valid kanban",
			vo:   &SaveProjectPreference{ViewType: "kanban", GetOptions: "all", Columns: "status"},
		},
		{
			name:    "empty view type",
			vo:      &SaveProjectPreference{ViewType: "", GetOptions: "all", Columns: "name"},
			wantErr: "视图类型不能为空",
		},
		{
			name:    "invalid view type",
			vo:      &SaveProjectPreference{ViewType: "gantt", GetOptions: "all", Columns: "name"},
			wantErr: "视图类型必须是 table、list 或 kanban",
		},
		{
			name:    "empty get options",
			vo:      &SaveProjectPreference{ViewType: "table", GetOptions: "", Columns: "name"},
			wantErr: "获取任务选项不能为空",
		},
		{
			name:    "empty columns",
			vo:      &SaveProjectPreference{ViewType: "table", GetOptions: "all", Columns: ""},
			wantErr: "列选项不能为空",
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
