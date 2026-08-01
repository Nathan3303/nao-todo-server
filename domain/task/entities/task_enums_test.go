package entities

import (
	"testing"
)

func TestParseTaskState(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TaskState
		wantW uint8
		wantK bool
	}{
		{name: "todo", input: "todo", want: TaskStatePending, wantW: 1, wantK: true},
		{name: "in-progress", input: "in-progress", want: TaskStateInProgress, wantW: 2, wantK: true},
		{name: "done", input: "done", want: TaskStateCompleted, wantW: 3, wantK: true},
		{name: "empty", input: "", want: 0, wantW: 0, wantK: false},
		{name: "unknown", input: "archived", want: 0, wantW: 0, wantK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseTaskState(tt.input)
			if ok != tt.wantK {
				t.Errorf("ParseTaskState(%q) ok = %v, want %v", tt.input, ok, tt.wantK)
			}
			if got != tt.want {
				t.Errorf("ParseTaskState(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if uint8(got) != tt.wantW {
				t.Errorf("ParseTaskState(%q) 底层值 = %d, want %d", tt.input, uint8(got), tt.wantW)
			}
		})
	}
}

func TestTaskState_String(t *testing.T) {
	tests := []struct {
		name  string
		state TaskState
		want  string
	}{
		{name: "pending", state: TaskStatePending, want: "todo"},
		{name: "in progress", state: TaskStateInProgress, want: "in-progress"},
		{name: "completed", state: TaskStateCompleted, want: "done"},
		{name: "zero value", state: TaskState(0), want: ""},
		{name: "out of range", state: TaskState(9), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("TaskState(%d).String() = %q, want %q", uint8(tt.state), got, tt.want)
			}
		})
	}
}

func TestTaskState_RoundTrip(t *testing.T) {
	for _, s := range []string{"todo", "in-progress", "done"} {
		t.Run(s, func(t *testing.T) {
			state, ok := ParseTaskState(s)
			if !ok {
				t.Fatalf("ParseTaskState(%q) ok = false, want true", s)
			}
			if got := state.String(); got != s {
				t.Errorf("往返结果 = %q, want %q", got, s)
			}
		})
	}
}

func TestParseTaskPriority(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  TaskPriority
		wantW uint8
		wantK bool
	}{
		{name: "low", input: "low", want: TaskPriorityLow, wantW: 1, wantK: true},
		{name: "medium", input: "medium", want: TaskPriorityMedium, wantW: 2, wantK: true},
		{name: "high", input: "high", want: TaskPriorityHigh, wantW: 3, wantK: true},
		{name: "urgent", input: "urgent", want: TaskPriorityUrgent, wantW: 4, wantK: true},
		{name: "none 无映射", input: "none", want: 0, wantW: 0, wantK: false},
		{name: "empty", input: "", want: 0, wantW: 0, wantK: false},
		{name: "unknown", input: "highest", want: 0, wantW: 0, wantK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseTaskPriority(tt.input)
			if ok != tt.wantK {
				t.Errorf("ParseTaskPriority(%q) ok = %v, want %v", tt.input, ok, tt.wantK)
			}
			if got != tt.want {
				t.Errorf("ParseTaskPriority(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if uint8(got) != tt.wantW {
				t.Errorf("ParseTaskPriority(%q) 底层值 = %d, want %d", tt.input, uint8(got), tt.wantW)
			}
		})
	}
}

func TestTaskPriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority TaskPriority
		want     string
	}{
		{name: "none", priority: TaskPriorityNone, want: ""},
		{name: "low", priority: TaskPriorityLow, want: "low"},
		{name: "medium", priority: TaskPriorityMedium, want: "medium"},
		{name: "high", priority: TaskPriorityHigh, want: "high"},
		{name: "urgent", priority: TaskPriorityUrgent, want: "urgent"},
		{name: "out of range", priority: TaskPriority(9), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.want {
				t.Errorf("TaskPriority(%d).String() = %q, want %q", uint8(tt.priority), got, tt.want)
			}
		})
	}
}

func TestTaskPriority_RoundTrip(t *testing.T) {
	for _, s := range []string{"low", "medium", "high", "urgent"} {
		t.Run(s, func(t *testing.T) {
			priority, ok := ParseTaskPriority(s)
			if !ok {
				t.Fatalf("ParseTaskPriority(%q) ok = false, want true", s)
			}
			if got := priority.String(); got != s {
				t.Errorf("往返结果 = %q, want %q", got, s)
			}
		})
	}
}

func TestParseRemindRepeat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  RemindRepeat
		wantW uint8
		wantK bool
	}{
		{name: "none", input: "none", want: RemindRepeatNone, wantW: 0, wantK: true},
		{name: "daily", input: "daily", want: RemindRepeatDaily, wantW: 1, wantK: true},
		{name: "weekly", input: "weekly", want: RemindRepeatWeekly, wantW: 2, wantK: true},
		{name: "monthly", input: "monthly", want: RemindRepeatMonthly, wantW: 3, wantK: true},
		{name: "empty", input: "", want: 0, wantW: 0, wantK: false},
		{name: "unknown", input: "yearly", want: 0, wantW: 0, wantK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseRemindRepeat(tt.input)
			if ok != tt.wantK {
				t.Errorf("ParseRemindRepeat(%q) ok = %v, want %v", tt.input, ok, tt.wantK)
			}
			if got != tt.want {
				t.Errorf("ParseRemindRepeat(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if uint8(got) != tt.wantW {
				t.Errorf("ParseRemindRepeat(%q) 底层值 = %d, want %d", tt.input, uint8(got), tt.wantW)
			}
		})
	}
}

func TestRemindRepeat_String(t *testing.T) {
	tests := []struct {
		name   string
		repeat RemindRepeat
		want   string
	}{
		{name: "none", repeat: RemindRepeatNone, want: "none"},
		{name: "daily", repeat: RemindRepeatDaily, want: "daily"},
		{name: "weekly", repeat: RemindRepeatWeekly, want: "weekly"},
		{name: "monthly", repeat: RemindRepeatMonthly, want: "monthly"},
		{name: "out of range", repeat: RemindRepeat(9), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.repeat.String(); got != tt.want {
				t.Errorf("RemindRepeat(%d).String() = %q, want %q", uint8(tt.repeat), got, tt.want)
			}
		})
	}
}

func TestRemindRepeat_RoundTrip(t *testing.T) {
	for _, s := range []string{"none", "daily", "weekly", "monthly"} {
		t.Run(s, func(t *testing.T) {
			repeat, ok := ParseRemindRepeat(s)
			if !ok {
				t.Fatalf("ParseRemindRepeat(%q) ok = false, want true", s)
			}
			if got := repeat.String(); got != s {
				t.Errorf("往返结果 = %q, want %q", got, s)
			}
		})
	}
}

func TestWeekdaysToBitmask(t *testing.T) {
	tests := []struct {
		name     string
		weekdays []uint8
		want     uint8
	}{
		{name: "empty", weekdays: []uint8{}, want: 0},
		{name: "nil", weekdays: nil, want: 0},
		{name: "sunday only", weekdays: []uint8{0}, want: 1},
		{name: "saturday only", weekdays: []uint8{6}, want: 64},
		{name: "all seven days", weekdays: []uint8{0, 1, 2, 3, 4, 5, 6}, want: 127},
		{name: "workdays", weekdays: []uint8{1, 2, 3, 4, 5}, want: 62},
		{name: "out of range skipped", weekdays: []uint8{1, 7, 200}, want: 2},
		{name: "duplicated", weekdays: []uint8{3, 3}, want: 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WeekdaysToBitmask(tt.weekdays); got != tt.want {
				t.Errorf("WeekdaysToBitmask(%v) = %d, want %d", tt.weekdays, got, tt.want)
			}
		})
	}
}

func TestBitmaskToWeekdays(t *testing.T) {
	tests := []struct {
		name string
		mask uint8
		want []uint8
	}{
		{name: "empty mask", mask: 0, want: []uint8{}},
		{name: "sunday only", mask: 1, want: []uint8{0}},
		{name: "saturday only", mask: 64, want: []uint8{6}},
		{name: "all seven days", mask: 127, want: []uint8{0, 1, 2, 3, 4, 5, 6}},
		{name: "workdays ascending", mask: 62, want: []uint8{1, 2, 3, 4, 5}},
		{name: "high bit ignored", mask: 128, want: []uint8{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BitmaskToWeekdays(tt.mask)
			if len(got) != len(tt.want) {
				t.Fatalf("BitmaskToWeekdays(%d) = %v, want %v", tt.mask, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("BitmaskToWeekdays(%d) = %v, want %v", tt.mask, got, tt.want)
				}
			}
		})
	}
}

func TestWeekdaysBitmask_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		weekdays []uint8
	}{
		{name: "empty", weekdays: []uint8{}},
		{name: "single value", weekdays: []uint8{4}},
		{name: "all seven days", weekdays: []uint8{0, 1, 2, 3, 4, 5, 6}},
		{name: "partial", weekdays: []uint8{0, 3, 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BitmaskToWeekdays(WeekdaysToBitmask(tt.weekdays))
			if len(got) != len(tt.weekdays) {
				t.Fatalf("往返结果 = %v, want %v", got, tt.weekdays)
			}
			for i := range got {
				if got[i] != tt.weekdays[i] {
					t.Fatalf("往返结果 = %v, want %v", got, tt.weekdays)
				}
			}
		})
	}
}
