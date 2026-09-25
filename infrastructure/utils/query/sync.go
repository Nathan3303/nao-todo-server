package query

import (
	"time"

	"gorm.io/gorm"
)

// SyncOrder 增量拉取稳定排序：updated_at 升序 + id 升序（二级排序）
// 保证游标"只前进不后退"，同 updated_at 时顺序确定
func SyncOrder() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("updated_at ASC").Order("id ASC")
	}
}

// ByKeysetCursor keyset 游标过滤：(updated_at, id) > (cursor, cursorID)
// 相比 offset 分页，分页期间的新写入不会造成重复或遗漏；
// cursor 为零值时不过滤（首次全量拉取）
func ByKeysetCursor(cursor time.Time, cursorID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if cursor.IsZero() {
			return db
		}
		return db.Where(
			"updated_at > ? OR (updated_at = ? AND id > ?)",
			cursor, cursor, cursorID,
		)
	}
}
