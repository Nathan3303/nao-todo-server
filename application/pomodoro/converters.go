package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/interfaces/types"
	"strconv"
	"time"
)

// CreatePomodoroRecordReqToVO 将 CreatePomodoroRecordReq 转换为 CreatePomodoroValueObject
func CreatePomodoroRecordReqToVO(
	userId int64,
	req *types.CreatePomodoroRecordReq,
) (*valueobjects.CreatePomodoroRecord, error) {
	taskId, err := strconv.ParseInt(req.TaskId, 10, 64)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreatePomodoroRecord(
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

// PomodoroRecordEntityToCreateRes 将 PomodoroRecordEntity 转换为 CreatePomodoroRecordRes
func PomodoroRecordEntityToCreateRes(e *entities.PomodoroRecord) *types.CreatePomodoroRecordRes {
	var res types.CreatePomodoroRecordRes
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

// PomodoroRecordEntityToGetRes 将 PomodoroRecordEntity 转换为 GetPomodoroRecordRes
func PomodoroRecordEntityToGetRes(e *entities.PomodoroRecord) *types.GetPomodoroRecordRes {
	var res types.GetPomodoroRecordRes
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

// PomodoroRecordEntitiesToGetReses 将 PomodoroRecordEntity 列表转换为 GetPomodoroRecordRes 列表
func PomodoroRecordEntitiesToGetReses(
	list []*entities.PomodoroRecord,
) []*types.GetPomodoroRecordRes {
	res := make([]*types.GetPomodoroRecordRes, 0, len(list))
	for _, e := range list {
		res = append(res, PomodoroRecordEntityToGetRes(e))
	}
	return res
}
