package query

import (
	"naotodoserver/infrastructure/utils"
	"strings"

	"gorm.io/gorm"
)

// Sort 通用排序 Scope
// 按 "field:direction" 格式解析排序参数，field 自动由驼峰转为下划线
func Sort(sort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort == "" {
			return db
		}
		parts := strings.Split(sort, ":")
		if len(parts) != 2 {
			return db
		}
		return db.Order(utils.ToSnakeCase(parts[0]) + " " + parts[1])
	}
}
