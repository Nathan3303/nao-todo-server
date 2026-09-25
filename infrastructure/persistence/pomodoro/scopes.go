package pomodoro

import (
	"time"

	"gorm.io/gorm"
)

// ByPomodoroRecordSessionId 按 session_id 精确匹配
func ByPomodoroRecordSessionId(sessionId string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sessionId == "" {
			return db
		}
		return db.Where("session_id = ?", sessionId)
	}
}

// ByPomodoroRecordTimeRange 按 start_at 时间范围过滤
// startTime/endTime 为 RFC3339 格式字符串，解析后用于 BETWEEN/>=/<= 查询
func ByPomodoroRecordTimeRange(startTime, endTime string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var startTimeParsed, endTimeParsed time.Time
		var hasStartTime, hasEndTime bool
		if startTime != "" {
			if t, err := time.Parse(time.RFC3339, startTime); err == nil {
				startTimeParsed = t
				hasStartTime = true
			}
		}
		if endTime != "" {
			if t, err := time.Parse(time.RFC3339, endTime); err == nil {
				endTimeParsed = t
				hasEndTime = true
			}
		}
		switch {
		case hasStartTime && hasEndTime:
			return db.Where("start_at BETWEEN ? AND ?", startTimeParsed, endTimeParsed)
		case hasStartTime:
			return db.Where("start_at >= ?", startTimeParsed)
		case hasEndTime:
			return db.Where("start_at <= ?", endTimeParsed)
		}
		return db
	}
}

// ByPomodoroRecordTaskId 按 task_id 精确匹配
func ByPomodoroRecordTaskId(taskId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if taskId <= 0 {
			return db
		}
		return db.Where("task_id = ?", taskId)
	}
}

// ByPomodoroRecordPomodoroId 按 pomodoro_id 精确匹配
func ByPomodoroRecordPomodoroId(pomodoroId int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pomodoroId <= 0 {
			return db
		}
		return db.Where("pomodoro_id = ?", pomodoroId)
	}
}

// ByPomodoroRecordTaskName 按 task_name 模糊匹配
func ByPomodoroRecordTaskName(taskName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if taskName == "" {
			return db
		}
		return db.Where("task_name LIKE ?", "%"+taskName+"%")
	}
}

// ByPomodoroType 按 type 精确匹配
func ByPomodoroType(pomodoroType uint8) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pomodoroType == 0 {
			return db
		}
		return db.Where("type = ?", pomodoroType)
	}
}

// ByPomodoroName 按 name 模糊匹配
func ByPomodoroName(name string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if name == "" {
			return db
		}
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

// ByPomodoroArchivedFlag 按归档状态过滤
// showArchived=true 时只显示已归档；false 时只显示未归档
func ByPomodoroArchivedFlag(isArchived bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isArchived {
			return db.Where("archived_at IS NOT NULL")
		}
		return db.Where("archived_at IS NULL")
	}
}
