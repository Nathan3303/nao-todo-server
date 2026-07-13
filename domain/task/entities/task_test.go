package entities

import (
	"testing"
	"time"

	"naotodoserver/domain/types"
)

func TestTask_IsEndAtValid(t *testing.T) {
	now := time.Now()
	future := now.Add(2 * time.Hour)
	past := now.Add(-2 * time.Hour)

	tests := []struct {
		name string
		task *Task
		want bool
	}{
		{
			name: "no end time",
			task: &Task{StartAt: types.NewNullableTimeByTime(now)},
			want: false,
		},
		{
			name: "end after start",
			task: &Task{
				StartAt: types.NewNullableTimeByTime(now),
				EndAt:   types.NewNullableTimeByTime(future),
			},
			want: true,
		},
		{
			name: "end before start",
			task: &Task{
				StartAt: types.NewNullableTimeByTime(now),
				EndAt:   types.NewNullableTimeByTime(past),
			},
			want: false,
		},
		{
			name: "no start time, has end time",
			task: &Task{
				EndAt: types.NewNullableTimeByTime(future),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.IsEndAtValid()
			if got != tt.want {
				t.Errorf("IsEndAtValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsDatesValid(t *testing.T) {
	now := time.Now()
	future := now.Add(2 * time.Hour)
	past := now.Add(-2 * time.Hour)

	tests := []struct {
		name    string
		task    *Task
		wantErr string
	}{
		{
			name: "all dates valid",
			task: &Task{
				StartAt:    types.NewNullableTimeByTime(now),
				EndAt:      types.NewNullableTimeByTime(future),
				ArchivedAt: types.NewNullableTimeNull(),
				StarMarkAt: types.NewNullableTimeNull(),
				GivenUpAt:  types.NewNullableTimeNull(),
			},
		},
		{
			name: "end before start",
			task: &Task{
				StartAt: types.NewNullableTimeByTime(now),
				EndAt:   types.NewNullableTimeByTime(past),
			},
			wantErr: "时间参数无效 - 结束时间必须晚于开始时间",
		},
		{
			name: "no dates set (both null = IsEndAtValid=false)",
			task: &Task{
				StartAt: types.NewNullableTimeNull(),
				EndAt:   types.NewNullableTimeNull(),
			},
			wantErr: "时间参数无效 - 结束时间必须晚于开始时间",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.IsDatesValid()
			if tt.wantErr == "" && err != nil {
				t.Errorf("IsDatesValid() unexpected error: %v", err)
			}
			if tt.wantErr != "" && err == nil {
				t.Errorf("IsDatesValid() want error %q, got nil", tt.wantErr)
			}
			if tt.wantErr != "" && err != nil && err.Error() != tt.wantErr {
				t.Errorf("IsDatesValid() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
