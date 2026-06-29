package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"
)

// CreatePomodoroVOToModel 转换创建待办任务番茄工作请求为数据库模型
func CreatePomodoroVOToModel(vo *valueobjects.CreatePomodoro) *models.Pomodoro {
	var m models.Pomodoro
	m.UserId = vo.UserId
	m.SessionId = vo.SessionId
	m.Type = vo.Type
	m.TaskId = vo.TaskId
	m.TaskName = vo.TaskName
	m.Description = vo.Description
	// vo.StartAt 和 vo.EndAt 经过校验，确保了非空指针
	m.StartAt = *vo.StartAt
	m.EndAt = *vo.EndAt
	m.Duration = vo.Duration
	m.Note = vo.Note
	return &m
}

// PomodoroModel2Entity 转换待办任务番茄工作模型为实体
func PomodoroModel2Entity(m *models.Pomodoro) *entities.Pomodoro {
	var e entities.Pomodoro
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = *types.NewNullableTimeWithTime(m.DeletedAt.Time)
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

// PomodoroModels2Entities 转换待办任务番茄工作模型列表为实体列表
func PomodoroModels2Entities(list []*models.Pomodoro) []*entities.Pomodoro {
	result := make([]*entities.Pomodoro, 0, len(list))
	for _, m := range list {
		result = append(result, PomodoroModel2Entity(m))
	}
	return result
}
