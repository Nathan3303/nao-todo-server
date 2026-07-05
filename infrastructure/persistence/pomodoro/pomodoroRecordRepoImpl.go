package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PomodoroRecordRepoImpl 实现 PomodoroRecordRepository 接口
type PomodoroRecordRepoImpl struct {
	db *gorm.DB
}

// NewPomodoroRecordRepo 创建 PomodoroRecordRepository 实例
func NewPomodoroRecordRepo(db *gorm.DB) repositories.PomodoroRecord {
	return &PomodoroRecordRepoImpl{db: db}
}

// Create 创建 PomodoroRecord
func (r *PomodoroRecordRepoImpl) Create(
	ctx context.Context,
	vo *valueobjects.CreatePomodoroRecord,
) (*entities.PomodoroRecord, error) {
	m := CreatePomodoroRecordVOToModel(vo)
	tx := r.db.WithContext(ctx).Create(m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return PomodoroRecordModel2Entity(m), nil
}

// GetById 获取 PomodoroRecord
func (r *PomodoroRecordRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	id int64,
) (*entities.PomodoroRecord, error) {
	var m models.PomodoroRecord
	tx := r.db.
		WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userId).
		First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return PomodoroRecordModel2Entity(&m), nil
}

// List 获取 PomodoroRecord 列表
func (r *PomodoroRecordRepoImpl) List(
	ctx context.Context,
	userId int64,
	sessionId, startTime, endTime string,
	taskId int64,
	taskName string,
	pomodoroType uint8,
	page, limit int,
	sort string,
) ([]*entities.PomodoroRecord, int64, error) {
	tx := r.db.WithContext(ctx).Model(&models.PomodoroRecord{}).
		Where("user_id = ?", userId)

	if sessionId != "" {
		tx = tx.Where("session_id = ?", sessionId)
	}

	// 解析 RFC3339 时间字符串为 time.Time，让 GORM 处理正确的 SQL 格式化
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
		tx = tx.Where("start_at BETWEEN ? AND ?", startTimeParsed, endTimeParsed)
	case hasStartTime:
		tx = tx.Where("start_at >= ?", startTimeParsed)
	case hasEndTime:
		tx = tx.Where("start_at <= ?", endTimeParsed)
	}
	if taskId > 0 {
		tx = tx.Where("task_id = ?", taskId)
	}
	if taskName != "" {
		tx = tx.Where("task_name LIKE ?", "%"+taskName+"%")
	}
	if pomodoroType > 0 {
		tx = tx.Where("type = ?", pomodoroType)
	}

	var total int64
	tx.Count(&total)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	if sort != "" {
		parts := strings.Split(sort, ":")
		if len(parts) == 2 {
			tx = tx.Order(utils.ToSnakeCase(parts[0]) + " " + parts[1])
		}
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	var modelsList []*models.PomodoroRecord
	tx = tx.Offset(offset).Limit(limit).Find(&modelsList)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	return PomodoroRecordModels2Entities(modelsList), total, nil
}
