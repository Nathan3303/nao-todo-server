package task

import (
	"naotodoserver/consts"
	"naotodoserver/domain/task/valueobjects"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ByProjectOrTag 按项目或标签过滤
// ProjectId > 0 时按精确项目匹配；否则按标签模糊匹配
func ByProjectOrTag(query *valueobjects.QueryTask) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if query.ProjectId > 0 {
			return db.Where("project_id = ?", query.ProjectId)
		} else if query.TagId != "" {
			return db.Where("tags LIKE ?", "%"+query.TagId+"%")
		}
		return db
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
			if val, exists := consts.TodoStateMap[s]; exists {
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
			if val, exists := consts.TodoPriorityMap[s]; exists {
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
			return db.Where("end_at >= ?", time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
		case "week":
			start, end := GetWeekRange(time.Now())
			return db.Where("end_at >= ? and end_at <= ?", start, end)
		case "month":
			return db.Where("end_at >= ?", time.Now().AddDate(0, 0, 7).Format("2006-01-02"))
		case "-today":
			return db.Where("end_at < ?", time.Now().Format("2006-01-02"))
		}
		return db
	}
}
