package tag

import (
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/vo"
	"naotodoserver/interfaces/types"
	"strconv"
)

func TagEntity2GetRes(e *entities.Tag) *types.GetTagRes {
	res := &types.GetTagRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.Name = e.Name
	res.Description = e.Description
	res.Color = e.Color
	res.Preference = TagPreferenceVO2Res(e.Preference)
	return res
}

func CreateReq2Entity(req *types.CreateTagReq) *entities.Tag {
	e := &entities.Tag{}
	e.Name = req.Name
	e.Description = req.Description
	e.Color = req.Color
	return e
}

func TagEntity2CreateRes(e *entities.Tag) *types.CreateTagRes {
	res := &types.CreateTagRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.Name = e.Name
	res.Description = e.Description
	res.Color = e.Color
	res.Preference = TagPreferenceVO2Res(e.Preference)
	return res
}

func TagPreferenceVO2Res(vo *vo.TagPreference) *types.TagPreferenceRes {
	if vo == nil {
		return nil
	}
	res := &types.TagPreferenceRes{}
	res.ViewType = vo.ViewType
	res.GetOptions = vo.GetOptions
	res.Columns = vo.Columns
	return res
}

func UpdateReq2Entity(req *types.UpdateTagReq) *entities.Tag {
	e := &entities.Tag{}
	e.Name = req.Name
	e.Description = req.Description
	e.Color = req.Color
	return e
}

func TagEntities2ListRes(eList []*entities.Tag) types.ListTagRes {
	res := types.ListTagRes{}
	for _, e := range eList {
		res = append(res, TagEntity2GetRes(e))
	}
	return res
}
