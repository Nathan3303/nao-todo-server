package idutil

import (
	"testing"
	"time"
)

// TestFormatTimeMilliRoundTrip 毫秒精度往返一致（游标精确性基础）
func TestFormatTimeMilliRoundTrip(t *testing.T) {
	ts := time.Date(2026, 1, 1, 10, 20, 30, 123456000, time.UTC)
	formatted := FormatTimeMilli(ts)
	want := "2026-01-01T10:20:30.123Z"
	if formatted != want {
		t.Fatalf("FormatTimeMilli = %q, want %q", formatted, want)
	}
	parsed, err := ParseTimeCompat(formatted)
	if err != nil {
		t.Fatalf("ParseTimeCompat(%q) error: %v", formatted, err)
	}
	if parsed.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("往返不一致: %v != %v", parsed.UnixMilli(), ts.UnixMilli())
	}
}

// TestParseTimeCompatLegacySeconds 兼容秒级 RFC3339（旧数据/旧客户端）
func TestParseTimeCompatLegacySeconds(t *testing.T) {
	parsed, err := ParseTimeCompat("2026-01-01T10:20:30Z")
	if err != nil {
		t.Fatalf("ParseTimeCompat 秒级格式 error: %v", err)
	}
	if parsed.Second() != 30 {
		t.Fatalf("秒级解析错误: %v", parsed)
	}
}
