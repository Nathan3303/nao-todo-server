package project

import (
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/vo"
	"naotodoserver/interfaces/types"
	"strconv"
)

func CreateReq2Entity(req *types.CreateProjectReq) *entities.Project {
	e := &entities.Project{}
	e.Name = req.Name
	e.Description = req.Description
	return e
}

func Entity2CreateRes(e *entities.Project) *types.CreateProjectRes {
	res := &types.CreateProjectRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.Name = e.Name
	res.ArchivedAt = e.ArchivedAt
	res.Description = e.Description
	res.Preference = PreferenceVO2Res(e.Preference)
	return res
}

func PreferenceVO2Res(e *vo.ProjectPreference) *types.ProjectPreferenceRes {
	if e == nil {
		return nil
	}
	res := &types.ProjectPreferenceRes{}
	res.ViewType = e.ViewType
	res.GetOptions = e.GetOptions
	res.Columns = e.Columns
	return res
}

func Entity2GetRes(e *entities.Project) *types.GetProjectRes {
	res := &types.GetProjectRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.Name = e.Name
	res.Description = e.Description
	res.ArchivedAt = e.ArchivedAt
	res.Preference = PreferenceVO2Res(e.Preference)
	return res
}

func UpdateReq2Entity(req *types.UpdateProjectReq) *entities.Project {
	e := &entities.Project{}
	e.Name = req.Name
	e.Description = req.Description
	return e
}

func Entities2ListRes(eList []*entities.Project) []*types.GetProjectRes {
	resList := make([]*types.GetProjectRes, 0, len(eList))
	for _, e := range eList {
		resList = append(resList, Entity2GetRes(e))
	}
	return resList
}
