package project

import "gorm.io/gorm"

// ByProjectArchived 按归档状态过滤（与 task.ByTaskArchived 同语义）
// isArchived=true 只返回已归档清单；false 与「未传」（Go 零值）一律返回未归档清单
// （archived_at IS NULL）——即服务端默认也排除归档（DP-1=(b)）。
// 注意：/sync/pull 走 ListSync，不经此 scope，归档清单照常同步（镜像完整）。
func ByProjectArchived(isArchived bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isArchived {
			return db.Where("archived_at IS NOT NULL")
		}
		return db.Where("archived_at IS NULL")
	}
}
