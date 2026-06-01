package utils

import (
	"database/sql"
	"time"

	"naotodoserver/domain/types"
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

func StringPtr2SqlNullTime(rfc3399DateStr *string) sql.NullTime {
	if rfc3399DateStr == nil {
		return sql.NullTime{Valid: false}
	}
	return String2SqlNullTime(*rfc3399DateStr)
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

func StringPtr2NullableTime(s *string) *types.NullableTime {
	if s == nil {
		return nil
	}
	if *s == "" || *s == "null" {
		return types.NewNullableTimeNull()
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return types.NewNullableTimeWithTime(t)
}

type NullableString interface {
	ToUpdateState() (shouldUpdate bool, isNull bool, value string)
}

func NullableString2NullableTime(ns NullableString) *types.NullableTime {
	shouldUpdate, isNull, value := ns.ToUpdateState()
	if !shouldUpdate {
		return nil
	}
	if isNull {
		return types.NewNullableTimeNull()
	}
	if value == "" {
		return types.NewNullableTimeNull()
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return types.NewNullableTimeWithTime(t)
}

// 获取某时间所在周的开始（周一）和结束（周日）
func GetWeekRange(t time.Time) (start, end time.Time) {
	// 将时间调整到 UTC 或本地时区（根据你的数据库时区设置）
	loc := time.Local // 或 time.UTC
	t = t.In(loc)

	// 计算距离周一的天数（Go 中 Weekday() 返回 0=Sunday, 1=Monday, ..., 6=Saturday）
	offset := int(t.Weekday())
	if offset == 0 {
		offset = 7 // 如果是周日，则往前推 6 天到周一
	}
	start = t.AddDate(0, 0, -offset+1) // 周一
	end = start.AddDate(0, 0, 6)       // 周日
	// 设置时间为当天的 00:00:00 到 23:59:59
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, end.Location())
	return
}
