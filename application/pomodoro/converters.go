package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
	"time"
)

func CreatePomodoroReqToVO(userId int64, req *types.CreatePomodoroReq) (*valueobjects.CreatePomodoro, error) {
	taskId, err := strconv.ParseInt(req.TaskId, 10, 64)
	if err != nil {
		return nil, err
	}

	startAt := utils.DateString2TimePtr(req.StartAt)
	endAt := utils.DateString2TimePtr(req.EndAt)

	return valueobjects.NewCreatePomodoro(
		userId,
		req.SessionId,
		req.Type,
		taskId,
		req.TaskName,
		req.Description,
		startAt,
		endAt,
		req.Duration,
		req.Note,
	)
}

func PomodoroEntityToCreateRes(e *entities.Pomodoro) *types.CreatePomodoroRes {
	return &types.CreatePomodoroRes{
		Id:          strconv.FormatInt(e.Id, 10),
		UserId:      strconv.FormatInt(e.UserId, 10),
		SessionId:   e.SessionId,
		Type:        e.Type,
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		TaskName:    e.TaskName,
		Description: e.Description,
		StartAt:     utils.Time2String(e.StartAt),
		EndAt:       utils.Time2String(e.EndAt),
		Duration:    e.Duration,
		Note:        e.Note,
		CreatedAt:   utils.Time2String(e.CreatedAt),
		UpdatedAt:   utils.Time2String(e.UpdatedAt),
	}
}

func PomodoroEntityToGetRes(e *entities.Pomodoro) *types.GetPomodoroRes {
	return &types.GetPomodoroRes{
		Id:          strconv.FormatInt(e.Id, 10),
		UserId:      strconv.FormatInt(e.UserId, 10),
		SessionId:   e.SessionId,
		Type:        e.Type,
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		TaskName:    e.TaskName,
		Description: e.Description,
		StartAt:     formatTime(e.StartAt),
		EndAt:       formatTime(e.EndAt),
		Duration:    e.Duration,
		Note:        e.Note,
		CreatedAt:   utils.Time2String(e.CreatedAt),
		UpdatedAt:   utils.Time2String(e.UpdatedAt),
		DeletedAt:   utils.Time2String(e.DeletedAt),
	}
}

func PomodoroEntitiesToGetReses(list []*entities.Pomodoro) []*types.GetPomodoroRes {
	res := make([]*types.GetPomodoroRes, 0, len(list))
	for _, e := range list {
		res = append(res, PomodoroEntityToGetRes(e))
	}
	return res
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
