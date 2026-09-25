package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// --- PomodoroRecord Converters ---

// CreatePomodoroRecordVOToModel 转换创建待办任务番茄工作记录为数据库模型
func CreatePomodoroRecordVOToModel(vo *valueobjects.CreatePomodoroRecord) *models.PomodoroRecord {
	var m models.PomodoroRecord
	m.ModelBase = models.ModelBase{
		ID:        vo.Id,
		CreatedAt: vo.CreatedAt,
		UpdatedAt: vo.UpdatedAt,
	}
	m.UserId = int64(vo.UserId)
	m.SessionId = vo.SessionId
	m.PomodoroId = int64(vo.PomodoroId)
	m.Type = uint8(vo.Type)
	m.TaskId = int64(vo.TaskId)
	m.TaskName = vo.TaskName
	m.Description = vo.Description
	// vo.StartAt 和 vo.EndAt 经过校验，确保了非空指针
	m.StartAt = vo.StartAt.Time
	m.EndAt = vo.EndAt.Time
	m.Duration = vo.Duration
	m.Note = vo.Note
	return &m
}

// PomodoroRecordModel2Entity 转换待办任务番茄工作记录模型为实体
func PomodoroRecordModel2Entity(m *models.PomodoroRecord) *entities.PomodoroRecord {
	var e entities.PomodoroRecord
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = types.UserID(m.UserId)
	e.PomodoroId = types.PomodoroID(m.PomodoroId)
	e.SessionId = m.SessionId
	e.Type = entities.PomodoroType(m.Type)
	e.TaskId = types.TaskID(m.TaskId)
	e.TaskName = m.TaskName
	e.Description = m.Description
	e.StartAt = m.StartAt
	e.EndAt = m.EndAt
	e.Duration = m.Duration
	e.Note = m.Note
	return &e
}

// PomodoroRecordModels2Entities 转换待办任务番茄工作记录模型列表为实体列表
func PomodoroRecordModels2Entities(list []*models.PomodoroRecord) []*entities.PomodoroRecord {
	result := make([]*entities.PomodoroRecord, 0, len(list))
	for _, m := range list {
		result = append(result, PomodoroRecordModel2Entity(m))
	}
	return result
}

// --- Pomodoro Converters ---

// CreatePomodoroVOToModel 转换创建常用番茄工作值对象为数据库模型
func CreatePomodoroVOToModel(vo *valueobjects.CreatePomodoro) *models.Pomodoro {
	var m models.Pomodoro
	m.ModelBase = models.ModelBase{
		ID:        vo.Id,
		CreatedAt: vo.CreatedAt,
		UpdatedAt: vo.UpdatedAt,
		DeletedAt: gorm.DeletedAt(vo.DeletedAt.ToSqlNullTime()),
	}
	m.UserId = int64(vo.UserId)
	m.Type = uint8(vo.Type)
	m.Name = vo.Name
	m.Description = vo.Description
	m.Duration = vo.Duration
	return &m
}

// CreatePomodoroVOToUpdateMap 创建常用番茄工作值对象转换为全量更新映射（Upsert 覆盖用）
// 仅包含 Create VO 表达的字段，不触碰 archived_at/total_duration 等列
// 客户端携带删除时间（本地墓碑）时写入 deleted_at；未携带时由 Upsert 兜底清空复活
func CreatePomodoroVOToUpdateMap(vo *valueobjects.CreatePomodoro) map[string]any {
	updateMap := map[string]any{
		"Type":        uint8(vo.Type),
		"Name":        vo.Name,
		"Description": vo.Description,
		"Duration":    vo.Duration,
	}
	if vo.DeletedAt.ShouldUpdate() && !vo.DeletedAt.IsSetToNull() {
		updateMap["deleted_at"] = vo.DeletedAt.ToSqlNullTime()
	}
	return updateMap
}

// CreatePomodoroRecordVOToUpdateMap 创建番茄工作记录值对象转换为全量更新映射（Upsert 覆盖用）
// 仅包含 Create VO 表达的字段，不触碰 archived_at
func CreatePomodoroRecordVOToUpdateMap(vo *valueobjects.CreatePomodoroRecord) map[string]any {
	return map[string]any{
		"SessionId":   vo.SessionId,
		"PomodoroId":  int64(vo.PomodoroId),
		"Type":        uint8(vo.Type),
		"TaskId":      int64(vo.TaskId),
		"TaskName":    vo.TaskName,
		"Description": vo.Description,
		"StartAt":     vo.StartAt.Time,
		"EndAt":       vo.EndAt.Time,
		"Duration":    vo.Duration,
		"Note":        vo.Note,
	}
}

// UpdatePomodoroVOToMap 转换更新常用番茄工作值对象为 map
// 仅包含需要更新的字段，实现 PATCH 语义
func UpdatePomodoroVOToMap(vo *valueobjects.UpdatePomodoro) map[string]any {
	updateMap := make(map[string]any)
	if vo.Type != nil {
		updateMap["Type"] = *vo.Type
	}
	if vo.Name != nil {
		updateMap["Name"] = *vo.Name
	}
	if vo.Description != nil {
		updateMap["Description"] = *vo.Description
	}
	if vo.Duration != nil {
		updateMap["Duration"] = *vo.Duration
	}
	if vo.ArchivedAt.ShouldUpdate() {
		if vo.ArchivedAt.IsSetToNull() {
			updateMap["ArchivedAt"] = nil
		} else {
			updateMap["ArchivedAt"] = vo.ArchivedAt.Time
		}
	}
	return updateMap
}

// PomodoroModel2Entity 转换常用番茄工作模型为实体
func PomodoroModel2Entity(m *models.Pomodoro) *entities.Pomodoro {
	var e entities.Pomodoro
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = types.UserID(m.UserId)
	e.Type = entities.PomodoroType(m.Type)
	e.Name = m.Name
	e.Description = m.Description
	e.Duration = m.Duration
	e.ArchivedAt = types.NewNullableTimeByTime(m.ArchivedAt.Time)
	e.TotalDuration = m.TotalDuration
	return &e
}

// PomodoroModels2Entities 转换常用番茄工作模型列表为实体列表
func PomodoroModels2Entities(list []*models.Pomodoro) []*entities.Pomodoro {
	result := make([]*entities.Pomodoro, 0, len(list))
	for _, m := range list {
		result = append(result, PomodoroModel2Entity(m))
	}
	return result
}
