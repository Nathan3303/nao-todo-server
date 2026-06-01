package utils

import (
	"database/sql"

	"naotodoserver/domain/types"
)

// ToSqlNullTime 将领域层 NullableTime 转换为 database/sql 的 NullTime
func ToSqlNullTime(nt *types.NullableTime) sql.NullTime {
	if nt == nil || !nt.ShouldUpdate() || nt.IsSetToNull() {
		return sql.NullTime{Valid: false}
	}
	time, _ := nt.Value()
	return sql.NullTime{Time: time, Valid: true}
}
