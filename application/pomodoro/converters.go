package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
	"time"
)

// CreatePomodoroReqToVO 将 CreatePomodoroReq 转换为 CreatePomodoroValueObject
func CreatePomodoroReqToVO(
	userId int64,
	req *types.CreatePomodoroReq,
) (*valueobjects.CreatePomodoro, error) {
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
		uint8(req.Duration),
		req.Note,
	)
}

// PomodoroEntityToCreateRes 将 PomodoroEntity 转换为 CreatePomodoroRes
func PomodoroEntityToCreateRes(e *entities.Pomodoro) *types.CreatePomodoroRes {
	return &types.CreatePomodoroRes{
		Id:          strconv.FormatInt(e.Id, 10),
		SessionId:   e.SessionId,
		Type:        e.Type,
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		TaskName:    e.TaskName,
		Description: e.Description,
		StartAt:     e.StartAt.Format(time.RFC3339),
		EndAt:       e.EndAt.Format(time.RFC3339),
		Duration:    int(e.Duration),
		Note:        e.Note,
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
	}
}

// PomodoroEntityToGetRes 将 PomodoroEntity 转换为 GetPomodoroRes
func PomodoroEntityToGetRes(e *entities.Pomodoro) *types.GetPomodoroRes {
	return &types.GetPomodoroRes{
		Id:          strconv.FormatInt(e.Id, 10),
		CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.Format(time.RFC3339),
		SessionId:   e.SessionId,
		Type:        e.Type,
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		TaskName:    e.TaskName,
		Description: e.Description,
		StartAt:     e.StartAt.Format(time.RFC3339),
		EndAt:       e.EndAt.Format(time.RFC3339),
		Duration:    int(e.Duration),
		Note:        e.Note,
	}
}

// PomodoroEntitiesToGetReses 将 PomodoroEntity 列表转换为 GetPomodoroRes 列表
func PomodoroEntitiesToGetReses(list []*entities.Pomodoro) []*types.GetPomodoroRes {
	res := make([]*types.GetPomodoroRes, 0, len(list))
	for _, e := range list {
		res = append(res, PomodoroEntityToGetRes(e))
	}
	return res
}
