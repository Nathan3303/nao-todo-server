package utils

import (
	"database/sql"
	"time"
)

func Time2String(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func DateString2Time(rfc3399DateStr string) time.Time {
	t, err := time.Parse(time.RFC3339, rfc3399DateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

func DateString2TimePtr(rfc3399DateStr string) *time.Time {
	t, err := time.Parse(time.RFC3339, rfc3399DateStr)
	if err != nil {
		return nil
	}
	return &t
}

func TimePtr2DateString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func TimePtr2SqlNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func String2SqlNullTime(rfc3399DateStr string) sql.NullTime {
	t, err := time.Parse(time.RFC3339, rfc3399DateStr)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func SqlNullTime2Time(t sql.NullTime) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func SqlNullTime2TimePtr(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
