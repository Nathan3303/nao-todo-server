package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
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
	return valueobjects.NewCreatePomodoro(
		userId,
		req.SessionId,
		req.Type,
		taskId,
		req.TaskName,
		req.Description,
		req.StartAt,
		req.EndAt,
		req.Duration,
		req.Note,
	)
}

// PomodoroEntityToCreateRes 将 PomodoroEntity 转换为 CreatePomodoroRes
func PomodoroEntityToCreateRes(e *entities.Pomodoro) *types.CreatePomodoroRes {
	var res types.CreatePomodoroRes
	res.Id = strconv.FormatInt(e.Id, 10)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.SessionId = e.SessionId
	res.Type = e.Type
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	res.TaskName = e.TaskName
	res.Description = e.Description
	res.StartAt = e.StartAt.Format(time.RFC3339)
	res.EndAt = e.EndAt.Format(time.RFC3339)
	res.Duration = e.Duration
	res.Note = e.Note
	return &res
}

// PomodoroEntityToGetRes 将 PomodoroEntity 转换为 GetPomodoroRes
func PomodoroEntityToGetRes(e *entities.Pomodoro) *types.GetPomodoroRes {
	var res types.GetPomodoroRes
	res.Id = strconv.FormatInt(e.Id, 10)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.SessionId = e.SessionId
	res.Type = e.Type
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	res.TaskName = e.TaskName
	res.Description = e.Description
	res.StartAt = e.StartAt.Format(time.RFC3339)
	res.EndAt = e.EndAt.Format(time.RFC3339)
	res.Duration = e.Duration
	res.Note = e.Note
	return &res
}

// PomodoroEntitiesToGetReses 将 PomodoroEntity 列表转换为 GetPomodoroRes 列表
func PomodoroEntitiesToGetReses(list []*entities.Pomodoro) []*types.GetPomodoroRes {
	res := make([]*types.GetPomodoroRes, 0, len(list))
	for _, e := range list {
		res = append(res, PomodoroEntityToGetRes(e))
	}
	return res
}
