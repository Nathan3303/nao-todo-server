package idutil

import "time"

// RFC3339Milli 毫秒精度的 RFC3339 时间格式
// 与数据库 datetime(3)（毫秒）精度一致；增量 keyset 游标依赖毫秒精度，
// 秒级截断会导致同秒记录被重复拉取
const RFC3339Milli = "2006-01-02T15:04:05.000Z07:00"

// FormatTimeMilli 格式化时间为毫秒精度 RFC3339
func FormatTimeMilli(t time.Time) string {
	return t.Format(RFC3339Milli)
}

// ParseTimeCompat 解析时间：优先毫秒格式，兼容秒级 RFC3339（旧数据/旧客户端）
func ParseTimeCompat(s string) (time.Time, error) {
	if t, err := time.Parse(RFC3339Milli, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
