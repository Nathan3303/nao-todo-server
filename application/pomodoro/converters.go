package pomodoro

import (
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/interfaces/types"
	"strconv"
	"time"
)

// --- PomodoroRecord Converters ---

// CreatePomodoroRecordReqToVO 将 CreatePomodoroRecordReq 转换为 CreatePomodoroValueObject
func CreatePomodoroRecordReqToVO(
	userId int64,
	req *types.CreatePomodoroRecordReq,
) (*valueobjects.CreatePomodoroRecord, error) {
	// TaskId 为弱关联，可为空（纯专注番茄）；仅在非空时解析
	var taskId int64
	var err error
	if req.TaskId != "" {
		taskId, err = strconv.ParseInt(req.TaskId, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	// PomodoroId 为弱关联，可为空；仅在非空时解析
	var pomodoroId int64
	if req.PomodoroId != "" {
		pomodoroId, err = strconv.ParseInt(req.PomodoroId, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return valueobjects.NewCreatePomodoroRecord(
		userId,
		req.SessionId,
		pomodoroId,
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
	res.PomodoroId = strconv.FormatInt(e.PomodoroId, 10)
	if e.PomodoroId == 0 {
		res.PomodoroId = ""
	}
	res.Type = e.Type
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	if e.TaskId == 0 {
		res.TaskId = ""
	}
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
	res.PomodoroId = strconv.FormatInt(e.PomodoroId, 10)
	if e.PomodoroId == 0 {
		res.PomodoroId = ""
	}
	res.Type = e.Type
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	if e.TaskId == 0 {
		res.TaskId = ""
	}
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

// ListPomodoroRecordReqToQueryVO 将 ListPomodoroRecordReq 转换为 QueryPomodoroRecord 值对象
func ListPomodoroRecordReqToQueryVO(
	userId int64,
	req *types.ListPomodoroRecordReq,
) *valueobjects.QueryPomodoroRecord {
	var taskId int64
	if req.TaskId != "" {
		var err error
		taskId, err = strconv.ParseInt(req.TaskId, 10, 64)
		if err != nil {
			taskId = 0
		}
	}
	var pomodoroId int64
	if req.PomodoroId != "" {
		var err error
		pomodoroId, err = strconv.ParseInt(req.PomodoroId, 10, 64)
		if err != nil {
			pomodoroId = 0
		}
	}
	return valueobjects.NewQueryPomodoroRecord(
		userId,
		pomodoroId,
		req.SessionId,
		req.StartTime,
		req.EndTime,
		taskId,
		req.TaskName,
		req.Type,
		req.Sort,
		req.Page,
		req.Limit,
	)
}

// ListPomodoroReqToQueryVO 将 ListPomodoroReq 转换为 QueryPomodoro 值对象
func ListPomodoroReqToQueryVO(
	userId int64,
	req *types.ListPomodoroReq,
) *valueobjects.QueryPomodoro {
	return valueobjects.NewQueryPomodoro(
		userId,
		req.Type,
		req.Name,
		req.IsArchived,
		req.Sort,
		req.Page,
		req.Limit,
	)
}

// --- Pomodoro Converters ---

// CreatePomodoroReqToVO 将 CreatePomodoroReq 转换为 CreatePomodoro 值对象
func CreatePomodoroReqToVO(
	userId int64,
	req *types.CreatePomodoroReq,
) (*valueobjects.CreatePomodoro, error) {
	return valueobjects.NewCreatePomodoro(
		userId,
		req.Type,
		req.Name,
		req.Description,
		req.Duration,
	)
}

// UpdatePomodoroReqToVO 将 UpdatePomodoroReq 转换为 UpdatePomodoro 值对象
func UpdatePomodoroReqToVO(
	req *types.UpdatePomodoroReq,
) (*valueobjects.UpdatePomodoro, error) {
	return valueobjects.NewUpdatePomodoro(
		req.Type,
		req.Name,
		req.Description,
		req.Duration,
		req.ArchivedAt,
	)
}

// PomodoroEntityToCreateRes 将 Pomodoro 实体转换为 CreatePomodoroRes
func PomodoroEntityToCreateRes(e *entities.Pomodoro) *types.CreatePomodoroRes {
	var res types.CreatePomodoroRes
	res.Id = strconv.FormatInt(e.Id, 10)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.Type = e.Type
	res.Name = e.Name
	res.Description = e.Description
	res.Duration = e.Duration
	res.ArchivedAt = e.ArchivedAt.ToString(time.RFC3339)
	res.TotalDuration = e.TotalDuration
	return &res
}

// PomodoroEntityToGetRes 将 Pomodoro 实体转换为 PomodoroRes
func PomodoroEntityToGetRes(e *entities.Pomodoro) *types.PomodoroRes {
	var res types.PomodoroRes
	res.Id = strconv.FormatInt(e.Id, 10)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.Type = e.Type
	res.Name = e.Name
	res.Description = e.Description
	res.Duration = e.Duration
	res.ArchivedAt = e.ArchivedAt.ToString(time.RFC3339)
	res.TotalDuration = e.TotalDuration
	return &res
}

// PomodoroEntitiesToGetReses 将 Pomodoro 实体列表转换为 PomodoroRes 列表
func PomodoroEntitiesToGetReses(
	list []*entities.Pomodoro,
) types.ListPomodoroRes {
	res := make(types.ListPomodoroRes, 0, len(list))
	for _, e := range list {
		res = append(res, *PomodoroEntityToGetRes(e))
	}
	return res
}
