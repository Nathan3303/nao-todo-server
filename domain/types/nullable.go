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

// NewNullableTimeWithTime 创建带时间值的 NullableTime
func NewNullableTimeWithTime(t time.Time) *NullableTime {
	return &NullableTime{
		Valid:  true,
		IsNull: false,
		Time:   t,
	}
}

// NewNullableTimeNull 创建空值 NullableTime
func NewNullableTimeNull() *NullableTime {
	return &NullableTime{
		Valid:  true,
		IsNull: true,
		Time:   time.Time{},
	}
}

// ShouldUpdate 是否需要更新
func (nt *NullableTime) ShouldUpdate() bool {
	return nt != nil && nt.Valid
}

// IsSetToNull 是否为设为空
func (nt *NullableTime) IsSetToNull() bool {
	return nt != nil && nt.Valid && nt.IsNull
}

// Value 获取时间值和是否有效，用于领域层时间比较
func (nt *NullableTime) Value() (time.Time, bool) {
	if nt == nil || !nt.Valid || nt.IsNull {
		return time.Time{}, false
	}
	return nt.Time, true
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
func (nt *NullableTime) ToString(fmt string) string {
	if nt == nil || !nt.ShouldUpdate() || nt.IsSetToNull() {
		return ""
	}
	if fmt == "" {
		fmt = time.RFC3339
	}
	return nt.Time.Format(fmt)
}
