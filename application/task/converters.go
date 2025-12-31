package task

import (
	"naotodoserver/consts"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/vo"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
	"time"
)

func TaskEntity2Res(e *entities.Task) *types.TaskRes {
	res := &types.TaskRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.ProjectId = strconv.FormatInt(e.ProjectId, 10)
	res.Name = e.Name
	res.Description = e.Description
	res.State = consts.TodoStateMapReverse[e.State]
	res.Priority = consts.TodoPriorityMapReverse[e.Priority]
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
	e.State = consts.TodoStateMap[req.State]
	e.Priority = consts.TodoPriorityMap[req.Priority]
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
	e.State = consts.TodoStateMap[req.State]
	e.Priority = consts.TodoPriorityMap[req.Priority]
	e.StartAt = utils.DateString2TimePtr(req.StartAt)
	e.EndAt = utils.DateString2TimePtr(req.EndAt)
	e.Tags = req.Tags
	if req.IsStarMarked {
		*e.StarMarkAt = time.Now()
	}
	return e
}

func ListTaskReq2QueryVO(req *types.ListTaskReq) *vo.TaskQuery {
	vo := &vo.TaskQuery{}
	vo.ProjectId, _ = strconv.ParseInt(req.ProjectId, 10, 64)
	vo.TagId = req.TagId
	vo.Name = req.Name
	vo.Description = req.Description
	vo.State = req.State
	vo.Priority = req.Priority
	vo.StartAt = req.StartAt
	vo.EndAt = req.EndAt
	vo.DeletedAt = req.DeletedAt
	vo.ArchivedAt = req.ArchivedAt
	vo.StarMarkAt = req.StarMarkAt
	vo.GivenUpAt = req.GivenUpAt
	vo.IsDeleted = req.IsDeleted
	vo.IsArchived = req.IsArchived
	vo.IsStarMarked = req.IsStarMarked
	vo.IsGivenUp = req.IsGivenUp
	vo.Page = req.Page
	vo.Limit = req.Limit
	vo.RelativeDate = req.RelativeDate
	vo.Sort = req.Sort
	return vo
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
