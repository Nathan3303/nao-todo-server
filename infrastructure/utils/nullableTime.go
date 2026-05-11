package utils

import (
	"database/sql"
	"time"
)

type NullableTime struct {
	Valid  bool
	IsNull bool
	Time   time.Time
}

func NewNullableTimeWithTime(t time.Time) *NullableTime {
	return &NullableTime{
		Valid:  true,
		IsNull: false,
		Time:   t,
	}
}

func NewNullableTimeNull() *NullableTime {
	return &NullableTime{
		Valid:  true,
		IsNull: true,
		Time:   time.Time{},
	}
}

func (nt *NullableTime) ToSqlNullTime() sql.NullTime {
	if nt == nil || !nt.Valid || nt.IsNull {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: nt.Time, Valid: true}
}

func (nt *NullableTime) ShouldUpdate() bool {
	return nt != nil && nt.Valid
}

func (nt *NullableTime) IsSetToNull() bool {
	return nt != nil && nt.Valid && nt.IsNull
}
