package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"
)

func CreatePomodoroVOToModel(vo *valueobjects.CreatePomodoro) *models.Pomodoro {
	return &models.Pomodoro{
		UserId:      vo.UserId,
		SessionId:   vo.SessionId,
		Type:        vo.Type,
		TaskId:      vo.TaskId,
		TaskName:    vo.TaskName,
		Description: vo.Description,
		StartAt:     utils.TimePtr2SqlNullTime(vo.StartAt),
		EndAt:       utils.TimePtr2SqlNullTime(vo.EndAt),
		Duration:    vo.Duration,
		Note:        vo.Note,
	}
}

func PomodoroModel2Entity(m *models.Pomodoro) *entities.Pomodoro {
	return &entities.Pomodoro{
		Id:          m.ID,
		UserId:      m.UserId,
		SessionId:   m.SessionId,
		Type:        m.Type,
		TaskId:      m.TaskId,
		TaskName:    m.TaskName,
		Description: m.Description,
		StartAt:     utils.SqlNullTime2Time(m.StartAt),
		EndAt:       utils.SqlNullTime2Time(m.EndAt),
		Duration:    m.Duration,
		Note:        m.Note,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt.Time,
	}
}

func PomodoroModels2Entities(list []*models.Pomodoro) []*entities.Pomodoro {
	result := make([]*entities.Pomodoro, 0, len(list))
	for _, m := range list {
		result = append(result, PomodoroModel2Entity(m))
	}
	return result
}
