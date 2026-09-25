package query

import "gorm.io/gorm"

// Paginate 分页 Scope
// 通用分页逻辑，从 PaginationVO2Scopes 迁移并通用化
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 10
		}
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}
