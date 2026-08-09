package pomodoro

import (
	"time"

	"naotodoserver/application/idutil"
	"naotodoserver/application/pomodoro/dto"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// --- PomodoroRecord Converters ---

// CreatePomodoroRecordReqToVO 将 CreatePomodoroRecordReq 转换为 CreatePomodoroValueObject
func CreatePomodoroRecordReqToVO(
	userId int64,
	req *dto.CreatePomodoroRecordReq,
) (*valueobjects.CreatePomodoroRecord, error) {
	// TaskId 为弱关联，可为空（纯专注番茄）；仅在非空时解析
	var taskId int64
	var err error
	if req.TaskId != "" {
		taskId, err = idutil.ParseID(req.TaskId)
		if err != nil {
			return nil, err
		}
	}
	// PomodoroId 为弱关联，可为空；仅在非空时解析
	var pomodoroId int64
	if req.PomodoroId != "" {
		pomodoroId, err = idutil.ParseID(req.PomodoroId)
		if err != nil {
			return nil, err
		}
	}
	vo, err := valueobjects.NewCreatePomodoroRecord(
		userId,
		req.SessionId,
		pomodoroId,
		entities.PomodoroType(req.Type),
		taskId,
		req.TaskName,
		req.Description,
		req.StartAt,
		req.EndAt,
		req.Duration,
		req.Note,
	)
	if err != nil {
		return nil, err
	}
	id, createdAt, updatedAt, err := idutil.ParseSyncMeta(req.Id, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	vo.Id = id
	vo.CreatedAt = createdAt
	vo.UpdatedAt = updatedAt
	return vo, nil
}

// PomodoroRecordEntityToCreateRes 将 PomodoroRecordEntity 转换为 CreatePomodoroRecordRes
func PomodoroRecordEntityToCreateRes(e *entities.PomodoroRecord) *dto.CreatePomodoroRecordRes {
	var res dto.CreatePomodoroRecordRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(idutil.RFC3339Milli)
	res.SessionId = e.SessionId
	res.PomodoroId = idutil.FormatID(e.PomodoroId)
	if e.PomodoroId == 0 {
		res.PomodoroId = ""
	}
	res.Type = uint8(e.Type)
	res.TaskId = idutil.FormatID(e.TaskId)
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
func PomodoroRecordEntityToGetRes(e *entities.PomodoroRecord) *dto.GetPomodoroRecordRes {
	var res dto.GetPomodoroRecordRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(idutil.RFC3339Milli)
	res.SessionId = e.SessionId
	res.PomodoroId = idutil.FormatID(e.PomodoroId)
	if e.PomodoroId == 0 {
		res.PomodoroId = ""
	}
	res.Type = uint8(e.Type)
	res.TaskId = idutil.FormatID(e.TaskId)
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
) []*dto.GetPomodoroRecordRes {
	res := make([]*dto.GetPomodoroRecordRes, 0, len(list))
	for _, e := range list {
		res = append(res, PomodoroRecordEntityToGetRes(e))
	}
	return res
}

// ListPomodoroRecordReqToQueryVO 将 ListPomodoroRecordReq 转换为 QueryPomodoroRecord 值对象
func ListPomodoroRecordReqToQueryVO(
	userId int64,
	req *dto.ListPomodoroRecordReq,
) *valueobjects.QueryPomodoroRecord {
	var taskId int64
	if req.TaskId != "" {
		var err error
		taskId, err = idutil.ParseID(req.TaskId)
		if err != nil {
			taskId = 0
		}
	}
	var pomodoroId int64
	if req.PomodoroId != "" {
		var err error
		pomodoroId, err = idutil.ParseID(req.PomodoroId)
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
		entities.PomodoroType(req.Type),
		req.Sort,
		req.Page,
		req.Limit,
	)
}

// ListPomodoroReqToQueryVO 将 ListPomodoroReq 转换为 QueryPomodoro 值对象
func ListPomodoroReqToQueryVO(
	userId int64,
	req *dto.ListPomodoroReq,
) *valueobjects.QueryPomodoro {
	return valueobjects.NewQueryPomodoro(
		userId,
		entities.PomodoroType(req.Type),
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
	req *dto.CreatePomodoroReq,
) (*valueobjects.CreatePomodoro, error) {
	vo, err := valueobjects.NewCreatePomodoro(
		userId,
		entities.PomodoroType(req.Type),
		req.Name,
		req.Description,
		req.Duration,
	)
	if err != nil {
		return nil, err
	}
	id, createdAt, updatedAt, err := idutil.ParseSyncMeta(req.Id, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	vo.Id = id
	vo.CreatedAt = createdAt
	vo.UpdatedAt = updatedAt
	return vo, nil
}

// UpdatePomodoroReqToVO 将 UpdatePomodoroReq 转换为 UpdatePomodoro 值对象
func UpdatePomodoroReqToVO(
	req *dto.UpdatePomodoroReq,
) (*valueobjects.UpdatePomodoro, error) {
	var voType *entities.PomodoroType
	if req.Type != nil {
		t := entities.PomodoroType(*req.Type)
		voType = &t
	}
	vo, err := valueobjects.NewUpdatePomodoro(
		voType,
		req.Name,
		req.Description,
		req.Duration,
		req.ArchivedAt,
	)
	if err != nil {
		return nil, err
	}
	if req.UpdatedAt != nil {
		t, err := idutil.ParseUpdatedAtCursor(*req.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vo.UpdatedAt = t
	}
	return vo, nil
}

// PomodoroEntityToCreateRes 将 Pomodoro 实体转换为 CreatePomodoroRes
func PomodoroEntityToCreateRes(e *entities.Pomodoro) *dto.CreatePomodoroRes {
	var res dto.CreatePomodoroRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(idutil.RFC3339Milli)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.Type = uint8(e.Type)
	res.Name = e.Name
	res.Description = e.Description
	res.Duration = e.Duration
	res.ArchivedAt = e.ArchivedAt.ToString(time.RFC3339)
	res.TotalDuration = e.TotalDuration
	return &res
}

// PomodoroEntityToGetRes 将 Pomodoro 实体转换为 PomodoroRes
func PomodoroEntityToGetRes(e *entities.Pomodoro) *dto.PomodoroRes {
	var res dto.PomodoroRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(idutil.RFC3339Milli)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.Type = uint8(e.Type)
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
) dto.ListPomodoroRes {
	res := make(dto.ListPomodoroRes, 0, len(list))
	for _, e := range list {
		res = append(res, *PomodoroEntityToGetRes(e))
	}
	return res
}
