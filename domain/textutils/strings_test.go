package textutils

import "testing"

func TestRuneLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty string", "", 0},
		{"ASCII string", "hello", 5},
		{"CJK characters", "你好世界", 4},
		{"mixed ASCII and CJK", "hello你好", 7},
		{"emoji", "👋🌍", 2},
		{"spaces", "a b c", 5},
		{"special chars", "abc!@#", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuneLength(tt.input)
			if got != tt.want {
				t.Errorf("RuneLength(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
