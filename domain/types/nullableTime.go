package types

import (
	"database/sql"
	"time"
)

// NullableTime 可空时间类型，用于 PATCH 语义：
// - nil: 未提供（不更新）
// - Valid=true, IsNull=true: 设为空
// - Valid=true, IsNull=false: 设为指定时间
type NullableTime struct {
	Valid  bool
	IsNull bool
	Time   time.Time
}

// ShouldUpdate 是否需要更新
func (nt *NullableTime) ShouldUpdate() bool {
	return nt != nil && nt.Valid
}

// IsSetToNull 是否为设为空值
func (nt *NullableTime) IsSetToNull() bool {
	return nt != nil && nt.Valid && nt.IsNull
}

// Value 获取时间值和是否有效，用于领域层时间比较
// @return (time.Time, bool) 时间值和是否有效
func (nt *NullableTime) Value() (time.Time, bool) {
	if nt == nil || !nt.Valid || nt.IsNull {
		return time.Time{}, false
	}
	return nt.Time, true
}

// SetTime 设置时间值
// @param t 时间值
func (nt *NullableTime) SetTime(t time.Time) {
	nt.Valid = true
	nt.IsNull = t.IsZero()
	nt.Time = t
}

// ToSqlNullTime 转换为 sql.NullTime
func (nt *NullableTime) ToSqlNullTime() sql.NullTime {
	if nt == nil || !nt.ShouldUpdate() || nt.IsSetToNull() {
		return sql.NullTime{Valid: false}
	}
	time, _ := nt.Value()
	return sql.NullTime{Time: time, Valid: true}
}

// ToString 转换为字符串
// @param fmt 时间格式
// @return string 时间字符串
func (nt *NullableTime) ToString(fmt string) string {
	if nt == nil || !nt.ShouldUpdate() || nt.IsSetToNull() || nt.IsNull {
		return ""
	}
	return nt.Time.Format(fmt)
}

// --- Maker ---

// NewNullableTimeNull 创建空值 NullableTime
// @return NullableTime
func NewNullableTimeNull() NullableTime {
	return NullableTime{
		Valid:  false,
		IsNull: true,
		Time:   time.Time{},
	}
}

// NewNullableTimeByTime 根据 time.Time 创建 NullableTime
// @param t 时间值
// @return NullableTime
func NewNullableTimeByTime(t time.Time) NullableTime {
	if t.IsZero() {
		return NullableTime{
			Valid:  false,
			IsNull: true,
			Time:   t,
		}
	}
	return NullableTime{
		Valid:  true,
		IsNull: false,
		Time:   t,
	}
}

// NewNullableTimeByTimePtr 根据 time.Time 指针创建 NullableTime
// @param t 时间指针
// @return NullableTime
func NewNullableTimeByTimePtr(tp *time.Time) NullableTime {
	if tp == nil {
		return NewNullableTimeNull()
	}
	return NewNullableTimeByTime(*tp)
}

// timeStrLayouts 支持解析的时间字符串格式，按优先级依次尝试
var timeStrLayouts = []string{
	time.RFC3339,       // 2006-01-02T15:04:05Z07:00
	"2006-01-02T15:04", // HTML datetime-local 格式
}

// NewNullableTimeByTimeStr 根据时间字符串创建 NullableTime
// @param ts 时间字符串
// @return NullableTime
func NewNullableTimeByTimeStr(ts string) NullableTime {
	if ts == "" {
		return NewNullableTimeNull()
	}
	for _, layout := range timeStrLayouts {
		if t, err := time.Parse(layout, ts); err == nil {
			return NewNullableTimeByTime(t)
		}
	}
	return NewNullableTimeNull()
}

// NewNullableTimeByTimeStrPtr 根据时间字符串指针创建 NullableTime
// @param ts 时间字符串指针
// @return NullableTime
func NewNullableTimeByTimeStrPtr(tp *string) NullableTime {
	if tp == nil {
		return NewNullableTimeNull()
	}
	return NewNullableTimeByTimeStr(*tp)
}
