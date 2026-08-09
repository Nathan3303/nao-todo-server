package idutil

import "time"

// ParseSyncMeta 解析客户端同步元数据（预置 id/createdAt/updatedAt）。
// 全部可空：nil 表示未提供——id 返回 0（由服务端雪花生成），时间返回零值（由 gorm 填充 now）。
// 提供但格式非法时返回错误，防止脏数据入库。
func ParseSyncMeta(id, createdAt, updatedAt *string) (int64, time.Time, time.Time, error) {
	var idVal int64
	if id != nil {
		v, err := ParseID(*id)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
		idVal = v
	}
	var created, updated time.Time
	if createdAt != nil {
		t, err := ParseTimeCompat(*createdAt)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
		created = t
	}
	if updatedAt != nil {
		t, err := ParseTimeCompat(*updatedAt)
		if err != nil {
			return 0, time.Time{}, time.Time{}, err
		}
		updated = t
	}
	return idVal, created, updated, nil
}

// ParseUpdatedAtCursor 解析增量同步游标（毫秒/秒级 RFC3339）；空字符串返回零值（不过滤）
func ParseUpdatedAtCursor(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return ParseTimeCompat(s)
}
