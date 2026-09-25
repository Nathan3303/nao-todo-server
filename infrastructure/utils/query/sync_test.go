package query

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestSyncScopesDryRun 用 gorm DryRun 验证增量 scope 生成的 SQL：
// keyset 游标过滤 (updated_at, id) > (cursor, cursorId)、稳定排序 updated_at ASC, id ASC
func TestSyncScopesDryRun(t *testing.T) {
	// 仅用于 DryRun 生成 SQL，不实际连接数据库
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "user:pass@tcp(127.0.0.1:3306)/db?parseTime=true",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}

	type stub struct {
		ID        int64 `gorm:"primaryKey"`
		UpdatedAt time.Time
	}

	t.Run("SyncOrder 生成稳定排序", func(t *testing.T) {
		stmt := db.Model(&stub{}).Scopes(SyncOrder()).Find(&stub{}).Statement
		sql := db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
		lower := strings.ToLower(sql)
		idxUpdated := strings.Index(lower, "updated_at asc")
		idxID := strings.Index(lower, "id asc")
		if idxUpdated < 0 || idxID < 0 || idxID < idxUpdated {
			t.Fatalf("排序应为 updated_at ASC, id ASC: %s", sql)
		}
	})

	t.Run("ByKeysetCursor 生成 (updated_at, id) 组合条件", func(t *testing.T) {
		cursor := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		stmt := db.Model(&stub{}).Scopes(ByKeysetCursor(cursor, 12345)).Find(&stub{}).Statement
		sql := db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
		lower := strings.ToLower(sql)
		if !strings.Contains(lower, "updated_at >") || !strings.Contains(lower, "id >") {
			t.Fatalf("keyset 应生成 (updated_at, id) 组合条件: %s", sql)
		}
	})

	t.Run("ByKeysetCursor 零值不过滤", func(t *testing.T) {
		stmt := db.Model(&stub{}).Scopes(ByKeysetCursor(time.Time{}, 0)).Find(&stub{}).Statement
		sql := db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...)
		if strings.Contains(sql, "updated_at") || strings.Contains(sql, "id") {
			t.Fatalf("零值 keyset 游标不应生成过滤条件: %s", sql)
		}
	})
}

// TestKeysetCursorSameSecond 同秒多记录场景下 keyset 推进不重复不遗漏：
// (updated_at, id) > (cursor, cursorId)，同秒内以 id 继续推进
func TestKeysetCursorSameSecond(t *testing.T) {
	type row struct {
		updatedAt time.Time
		id        int64
	}
	// 同一秒内 5 条记录
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	all := make([]row, 0, 5)
	for id := int64(101); id <= 105; id++ {
		all = append(all, row{updatedAt: ts, id: id})
	}

	// 模拟逐页 keyset 拉取：每页取 (updated_at, id) > (cursor, cursorID) 的前 limit 条
	collect := func(cursor time.Time, cursorID int64, limit int) []row {
		var out []row
		for _, r := range all {
			if r.updatedAt.After(cursor) || (r.updatedAt.Equal(cursor) && r.id > cursorID) {
				out = append(out, r)
				if len(out) == limit {
					break
				}
			}
		}
		return out
	}

	var got []row
	cursor, cursorID := time.Time{}, int64(0)
	for {
		page := collect(cursor, cursorID, 2)
		if len(page) == 0 {
			break
		}
		got = append(got, page...)
		last := page[len(page)-1]
		cursor, cursorID = last.updatedAt, last.id
		if len(got) > len(all) {
			t.Fatalf("keyset 推进出现重复: %+v", got)
		}
	}

	if len(got) != len(all) {
		t.Fatalf("keyset 推进遗漏: got %d, want %d (%+v)", len(got), len(all), got)
	}
	for i, r := range got {
		if r.id != all[i].id {
			t.Fatalf("顺序不一致: got %+v, want %+v", got, all)
		}
	}
}
