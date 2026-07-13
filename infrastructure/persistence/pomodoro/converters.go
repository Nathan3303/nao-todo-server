package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"
)

// --- PomodoroRecord Converters ---

// CreatePomodoroRecordVOToModel 转换创建待办任务番茄工作记录为数据库模型
func CreatePomodoroRecordVOToModel(vo *valueobjects.CreatePomodoroRecord) *models.PomodoroRecord {
	var m models.PomodoroRecord
	m.UserId = vo.UserId
	m.SessionId = vo.SessionId
	m.PomodoroId = vo.PomodoroId
	m.Type = uint8(vo.Type)
	m.TaskId = vo.TaskId
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
	e.UserId = m.UserId
	e.PomodoroId = m.PomodoroId
	e.SessionId = m.SessionId
	e.Type = entities.PomodoroType(m.Type)
	e.TaskId = m.TaskId
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
	m.UserId = vo.UserId
	m.Type = uint8(vo.Type)
	m.Name = vo.Name
	m.Description = vo.Description
	m.Duration = vo.Duration
	return &m
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
	e.UserId = m.UserId
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
