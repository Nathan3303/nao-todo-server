package event

import (
	"naotodoserver/domain/event/entities"
	"naotodoserver/interfaces/types"
	"strconv"
)

func EventEntity2Res(e *entities.Event) *types.EventRes {
	res := &types.EventRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	res.Name = e.Name
	res.Description = e.Description
	res.IsDone = e.IsDone
	res.SortId = e.SortId
	return res
}

func CreateEventReq2Entity(req *types.CreateEventReq) *entities.Event {
	entity := &entities.Event{}
	entity.TaskId, _ = strconv.ParseInt(req.TaskId, 10, 64)
	entity.Name = req.Name
	entity.Description = req.Description
	return entity
}

func UpdateEventReq2Entity(req *types.UpdateEventReq) *entities.Event {
	entity := &entities.Event{}
	entity.Name = req.Name
	entity.Description = req.Description
	entity.IsDone = req.IsDone
	entity.SortId = req.SortId
	return entity
}

func EventEntities2Reses(eList []*entities.Event) types.ListEventRes {
	reses := make([]*types.EventRes, 0, len(eList))
	for _, e := range eList {
		reses = append(reses, EventEntity2Res(e))
	}
	return reses
}
