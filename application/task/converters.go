package task

import (
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/vo"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
	"time"
)

var TodoStateMap = map[string]int8{
	"todo":        1,
	"in-progress": 2,
	"done":        3,
}

var TodoStateMapReverse = map[int8]string{
	1: "todo",
	2: "in-progress",
	3: "done",
}

var TodoPriorityMap = map[string]int8{
	"low":    1,
	"medium": 2,
	"high":   3,
	"urgent": 4,
}

var TodoPriorityMapReverse = map[int8]string{
	1: "low",
	2: "medium",
	3: "high",
	4: "urgent",
}

func TaskEntity2Res(e *entities.Task) *types.TaskRes {
	res := &types.TaskRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.ProjectId = strconv.FormatInt(e.ProjectId, 10)
	res.Name = e.Name
	res.Description = e.Description
	res.State = TodoStateMapReverse[e.State]
	res.Priority = TodoPriorityMapReverse[e.Priority]
	res.StartAt = e.GetFormatedStartAt()
	res.EndAt = e.GetFormatedEndAt()
	res.ArchivedAt, res.IsArchived = e.ParseArchivedAt()
	res.StarMarkAt, res.IsStarMarked = e.ParseStarMarkAt()
	res.GivenUpAt, res.IsGivenUp = e.ParseGivenUpAt()
	res.Tags = e.Tags
	res.UpdatedAt = e.GetFormatedUpdatedAt()
	res.CreatedAt = e.GetFormatedCreatedAt()
	res.DeletedAt, res.IsDeleted = e.GetFormatedDeletedAt()
	return res
}

func CreateTaskReq2Entity(req *types.CreateTaskReq) *entities.Task {
	e := &entities.Task{}
	e.ProjectId, _ = strconv.ParseInt(req.ProjectId, 10, 64)
	e.Name = req.Name
	e.Description = req.Description
	e.State = TodoStateMap[req.State]
	e.Priority = TodoPriorityMap[req.Priority]
	e.StartAt = utils.DateString2TimePtr(req.StartAt)
	e.EndAt = utils.DateString2TimePtr(req.EndAt)
	e.Tags = req.Tags
	return e
}

func UpdateTaskReq2Entity(req *types.UpdateTaskReq) *entities.Task {
	e := &entities.Task{}
	e.ProjectId, _ = strconv.ParseInt(req.ProjectId, 10, 64)
	e.Name = req.Name
	e.Description = req.Description
	e.State = TodoStateMap[req.State]
	e.Priority = TodoPriorityMap[req.Priority]
	e.StartAt = utils.DateString2TimePtr(req.StartAt)
	e.EndAt = utils.DateString2TimePtr(req.EndAt)
	e.Tags = req.Tags
	if req.IsStarMarked {
		*e.StarMarkAt = time.Now()
	}
	return e
}

func ListTaskReq2Entity(req *types.ListTaskReq) *entities.Task {
	e := &entities.Task{}
	e.ProjectId, _ = strconv.ParseInt(req.ProjectId, 10, 64)
	e.Name = req.Name
	e.Description = req.Description
	e.State = TodoStateMap[req.State]
	e.Priority = TodoPriorityMap[req.Priority]
	e.StartAt = utils.DateString2TimePtr(req.StartAt)
	e.EndAt = utils.DateString2TimePtr(req.EndAt)
	e.DeletedAt = utils.DateString2Time(req.DeletedAt)
	return e
}

func TaskEntities2Reses(e []*entities.Task) []*types.TaskRes {
	reses := make([]*types.TaskRes, 0, len(e))
	for _, item := range e {
		reses = append(reses, TaskEntity2Res(item))
	}
	return reses
}

func PaginationVO2Res(pagination *vo.Pagination) *types.Pagination {
	res := &types.Pagination{}
	res.Page = pagination.Page
	res.Limit = pagination.Limit
	res.Total = pagination.Total
	res.MaxPage = int(pagination.Total / int64(pagination.Limit))
	if pagination.Total%int64(pagination.Limit) > 0 {
		res.MaxPage++
	}
	return res
}
