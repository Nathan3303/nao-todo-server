package task

import (
	"naotodoserver/domain/task/entities"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ByParentTaskId 按父任务 ID 过滤
// parentTaskId > 0 时按精确匹配；parentTaskId == 0 时不筛选
func ByParentTaskId(parentTaskId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if parentTaskId <= 0 {
			// return db
			return db.Where("parent_task_id = 0")
		}
		return db.Where("parent_task_id = ?", parentTaskId)
	}
}

// ByProjects 按项目过滤（多值组内 OR）：project_id IN (ids)；空集不过滤。
func ByProjects(projectIds []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(projectIds) == 0 {
			return db
		}
		return db.Where("project_id IN ?", projectIds)
	}
}

// ByTags 按标签过滤（多值组内 OR）：tags JSON 串任一命中子串即算；空集不过滤。
// 与既有单值语义一致（LIKE 子串，非 JSON_CONTAINS 精确匹配）。
func ByTags(tagIds []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(tagIds) == 0 {
			return db
		}
		conds := make([]string, 0, len(tagIds))
		args := make([]any, 0, len(tagIds))
		for _, id := range tagIds {
			conds = append(conds, "tags LIKE ?")
			args = append(args, "%"+id+"%")
		}
		expr := strings.Join(conds, " OR ")
		if len(conds) > 1 {
			// 组内 OR 显式加括号，保证与其它过滤（组间 AND）组合时原子
			expr = "(" + expr + ")"
		}
		return db.Where(expr, args...)
	}
}

// ByTaskName 按名称模糊匹配
func ByTaskName(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if name == "" {
			return db
		}
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

// ByTaskDescription 按描述模糊匹配
func ByTaskDescription(desc string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if desc == "" {
			return db
		}
		return db.Where("description LIKE ?", "%"+desc+"%")
	}
}

// ByTaskState 按状态过滤
// 支持逗号分隔的多个状态（如 "todo,in_progress"）
func ByTaskState(state string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if state == "" {
			return db
		}
		var stateIDs []int
		for s := range strings.SplitSeq(state, ",") {
			if val, ok := entities.ParseTaskState(s); ok {
				stateIDs = append(stateIDs, int(val))
			}
		}
		if len(stateIDs) > 0 {
			return db.Where("state IN ?", stateIDs)
		}
		return db
	}
}

// ByTaskPriority 按优先级过滤
// 支持逗号分隔的多个优先级（如 "low,medium"）
func ByTaskPriority(priority string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if priority == "" {
			return db
		}
		var priorityIDs []int
		for s := range strings.SplitSeq(priority, ",") {
			if val, ok := entities.ParseTaskPriority(s); ok {
				priorityIDs = append(priorityIDs, int(val))
			}
		}
		if len(priorityIDs) > 0 {
			return db.Where("priority IN ?", priorityIDs)
		}
		return db
	}
}

// ByTaskTimeRange 按开始/结束时间范围过滤
func ByTaskTimeRange(startAt, endAt string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if startAt != "" {
			db = db.Where("start_at >= ?", startAt)
		}
		if endAt != "" {
			db = db.Where("end_at <= ?", endAt)
		}
		return db
	}
}

// ByTaskDeleted 查询已删除的任务（软删除）
func ByTaskDeleted(isDeleted bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isDeleted {
			return db.Unscoped().Where("deleted_at IS NOT NULL")
		}
		return db
	}
}

// ByTaskArchived 按归档状态过滤
func ByTaskArchived(isArchived bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isArchived {
			return db.Where("archived_at IS NOT NULL")
		}
		return db
	}
}

// ByTaskStarMarked 按星标状态过滤
func ByTaskStarMarked(isStarMarked bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isStarMarked {
			return db.Where("star_mark_at IS NOT NULL")
		}
		return db
	}
}

// ByTaskGivenUpFlag 按放弃状态过滤
// showGivenUp=true 时只显示已放弃的任务；false 时排除已放弃的任务
func ByTaskGivenUpFlag(showGivenUp bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if showGivenUp {
			return db.Where("given_up_at IS NOT NULL")
		}
		return db.Where("given_up_at IS NULL")
	}
}

// ByRelativeDate 按相对日期过滤
func ByRelativeDate(relativeDate string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if relativeDate == "" {
			return db
		}
		switch relativeDate {
		case "today":
			return db.Where("end_at >= ?", time.Now().Format("2006-01-02"))
		case "tomorrow":
			return db.Where(
				"end_at >= ?",
				time.
					Now().
					AddDate(0, 0, 1).
					Format("2006-01-02"),
			)
		case "week":
			start, end := GetWeekRange(time.Now())
			return db.Where("end_at >= ? and end_at <= ?", start, end)
		case "month":
			// 真实自然月窗口（本地时区）：end_at ∈ [本月1日, 次月1日)
			now := time.Now().In(time.Local)
			y, m, _ := now.Date()
			start := time.Date(y, m, 1, 0, 0, 0, 0, time.Local)
			return db.Where("end_at >= ? AND end_at < ?", start, start.AddDate(0, 1, 0))
		case "-today":
			return db.Where("end_at < ?", time.Now().Format("2006-01-02"))
		}
		return db
	}
}
