package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"
)

// CreatePomodoroRecordVOToModel 转换创建待办任务番茄工作记录为数据库模型
func CreatePomodoroRecordVOToModel(vo *valueobjects.CreatePomodoroRecord) *models.PomodoroRecord {
	var m models.PomodoroRecord
	m.UserId = vo.UserId
	m.SessionId = vo.SessionId
	m.Type = vo.Type
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
	e.SessionId = m.SessionId
	e.Type = m.Type
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
