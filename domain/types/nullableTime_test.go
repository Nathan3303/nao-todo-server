package types

import (
	"testing"
	"time"
)

func TestNewNullableTimeNull(t *testing.T) {
	nt := NewNullableTimeNull()
	if nt.Valid {
		t.Error("NewNullableTimeNull().Valid = true, want false")
	}
	if !nt.IsNull {
		t.Error("NewNullableTimeNull().IsNull = false, want true")
	}
}

func TestNewNullableTimeByTime(t *testing.T) {
	now := time.Now()
	nt := NewNullableTimeByTime(now)
	if !nt.Valid {
		t.Error("NewNullableTimeByTime(non-zero).Valid = false, want true")
	}
	if nt.IsNull {
		t.Error("NewNullableTimeByTime(non-zero).IsNull = true, want false")
	}
	if !nt.Time.Equal(now) {
		t.Errorf("NewNullableTimeByTime(non-zero).Time = %v, want %v", nt.Time, now)
	}

	// zero time
	zero := time.Time{}
	nt2 := NewNullableTimeByTime(zero)
	if nt2.Valid {
		t.Error("NewNullableTimeByTime(zero).Valid = true, want false")
	}
	if !nt2.IsNull {
		t.Error("NewNullableTimeByTime(zero).IsNull = false, want true")
	}

	// zero time with explicit IsNull
	nt3 := NewNullableTimeByTime(time.Time{})
	if nt3.Valid {
		t.Error("NewNullableTimeByTime(zero).Valid should be false")
	}
}

func TestNewNullableTimeByTimePtr(t *testing.T) {
	now := time.Now()
	nt := NewNullableTimeByTimePtr(&now)
	if !nt.Valid || nt.IsNull {
		t.Error("NewNullableTimeByTimePtr(non-nil) should be valid and non-null")
	}

	nt2 := NewNullableTimeByTimePtr(nil)
	if nt2.Valid {
		t.Error("NewNullableTimeByTimePtr(nil).Valid = true, want false")
	}
}

func TestNewNullableTimeByTimeStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNull bool // IsNull == true means the value is null
	}{
		{"RFC3339 valid", "2024-01-15T10:30:00Z", false},
		{"HTML datetime-local", "2024-01-15T10:30", false},
		{"empty string", "", true},
		{"invalid format", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nt := NewNullableTimeByTimeStr(tt.input)
			if nt.Valid != true {
				t.Errorf("NewNullableTimeByTimeStr(%q).Valid = false, want true (null is still Valid)", tt.input)
			}
			if nt.IsNull != tt.wantNull {
				t.Errorf("NewNullableTimeByTimeStr(%q).IsNull = %v, want %v", tt.input, nt.IsNull, tt.wantNull)
			}
		})
	}
}

func TestNewNullableTimeByTimeStrPtr(t *testing.T) {
	ts := "2024-01-15T10:30:00Z"
	nt := NewNullableTimeByTimeStrPtr(&ts)
	if !nt.Valid || nt.IsNull {
		t.Error("NewNullableTimeByTimeStrPtr(non-nil valid string) should be valid")
	}

	nt2 := NewNullableTimeByTimeStrPtr(nil)
	if nt2.Valid {
		t.Error("NewNullableTimeByTimeStrPtr(nil).Valid = true, want false")
	}
}

func TestNullableTime_ShouldUpdate(t *testing.T) {
	tests := []struct {
		name string
		nt   *NullableTime
		want bool
	}{
		{"nil pointer", nil, false},
		{"Valid=true, IsNull=false", &NullableTime{Valid: true, IsNull: false}, true},
		{"Valid=true, IsNull=true", &NullableTime{Valid: true, IsNull: true}, true},
		{"Valid=false", &NullableTime{Valid: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.nt.ShouldUpdate()
			if got != tt.want {
				t.Errorf("ShouldUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullableTime_IsSetToNull(t *testing.T) {
	tests := []struct {
		name string
		nt   *NullableTime
		want bool
	}{
		{"nil pointer", nil, false},
		{"Valid=true, IsNull=true", &NullableTime{Valid: true, IsNull: true}, true},
		{"Valid=true, IsNull=false", &NullableTime{Valid: true, IsNull: false}, false},
		{"Valid=false", &NullableTime{Valid: false, IsNull: true}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.nt.IsSetToNull()
			if got != tt.want {
				t.Errorf("IsSetToNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullableTime_Value(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		nt      *NullableTime
		want    time.Time
		wantOk  bool
	}{
		{"nil", nil, time.Time{}, false},
		{"valid non-null", &NullableTime{Valid: true, IsNull: false, Time: now}, now, true},
		{"valid null", &NullableTime{Valid: true, IsNull: true}, time.Time{}, false},
		{"invalid", &NullableTime{Valid: false}, time.Time{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.nt.Value()
			if ok != tt.wantOk {
				t.Errorf("Value() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && !got.Equal(tt.want) {
				t.Errorf("Value() time = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullableTime_SetTime(t *testing.T) {
	now := time.Now()
	var nt NullableTime
	nt.SetTime(now)
	if !nt.Valid {
		t.Error("SetTime(non-zero).Valid = false, want true")
	}
	if nt.IsNull {
		t.Error("SetTime(non-zero).IsNull = true, want false")
	}
	if !nt.Time.Equal(now) {
		t.Errorf("SetTime(non-zero).Time = %v, want %v", nt.Time, now)
	}

	// Set zero time
	nt.SetTime(time.Time{})
	if !nt.Valid {
		t.Error("SetTime(zero).Valid = false, want true")
	}
	if !nt.IsNull {
		t.Error("SetTime(zero).IsNull = false, want true")
	}
}

func TestNullableTime_ToString(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	tests := []struct {
		name string
		nt   *NullableTime
		want string
	}{
		{"nil", nil, ""},
		{"valid non-null", &NullableTime{Valid: true, IsNull: false, Time: now}, "2024-01-15T10:30:00Z"},
		{"valid null", &NullableTime{Valid: true, IsNull: true}, ""},
		{"invalid", &NullableTime{Valid: false}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.nt.ToString(time.RFC3339)
			if got != tt.want {
				t.Errorf("ToString() = %q, want %q", got, tt.want)
			}
		})
	}
}
