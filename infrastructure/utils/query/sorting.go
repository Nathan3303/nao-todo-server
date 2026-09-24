package query

import (
	"naotodoserver/infrastructure/utils"
	"strings"

	"gorm.io/gorm"
)

// sortableFields 排序字段白名单（task / pomodoro / pomodoro_record 既有排序列的并集），
// 键为落库后的 snake_case 列名。输入字段先经 ToSnakeCase 归一（兼容驼峰与下划线写法），
// 未命中白名单一律忽略（回落调用方默认序），避免 ORDER BY 字段原样拼接形成注入面。
var sortableFields = map[string]struct{}{
	// 通用列
	"id": {}, "created_at": {}, "updated_at": {}, "deleted_at": {},
	// task 列
	"parent_task_id": {}, "project_id": {}, "name": {}, "description": {},
	"state": {}, "priority": {}, "tags": {}, "start_at": {}, "end_at": {},
	"archived_at": {}, "star_mark_at": {}, "given_up_at": {}, "completed_at": {},
	"remind_at": {}, "sort_id": {},
	// pomodoro 列
	"type": {}, "duration": {}, "total_duration": {},
	// pomodoro_record 列
	"pomodoro_id": {}, "session_id": {}, "task_id": {}, "task_name": {}, "note": {},
}

// sortDirections 排序方向白名单（仅 asc / desc，大小写不敏感）
var sortDirections = map[string]struct{}{
	"asc": {}, "desc": {},
}

// Sort 通用排序 Scope
// 按 "field:direction" 格式解析排序参数，field 自动由驼峰转为下划线。
// 字段与方向均须命中白名单；非法输入一律忽略（回落调用方默认序），不参与 SQL 拼接。
func Sort(sort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort == "" {
			return db
		}
		parts := strings.Split(sort, ":")
		if len(parts) != 2 {
			return db
		}
		column := utils.ToSnakeCase(parts[0])
		direction := strings.ToLower(parts[1])
		if _, ok := sortableFields[column]; !ok {
			return db
		}
		if _, ok := sortDirections[direction]; !ok {
			return db
		}
		return db.Order(column + " " + direction)
	}
}
