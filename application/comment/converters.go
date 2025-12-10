package comment

import (
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/vo"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
)

func CommentEntity2Res(e *entities.Comment) *types.CommentRes {
	res := &types.CommentRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	res.Content = e.Content
	res.CreatedAt = utils.TimePtr2DateString(e.CreatedAt)
	res.Attachments = e.Attachments
	res.IsTopUp = e.IsTopUp
	res.CommentUser = CommentUserVO2Res(e.CommentUser)
	return res
}

func CommentUserVO2Res(vo *vo.CommentUser) *types.CommentUserRes {
	res := &types.CommentUserRes{}
	res.Avatar = vo.Avatar
	res.Nickname = vo.Nickname
	return res
}

func CreateCommentReq2Entity(req *types.CreateCommentReq) *entities.Comment {
	entity := &entities.Comment{}
	entity.TaskId, _ = strconv.ParseInt(req.TaskId, 10, 64)
	entity.Content = req.Content
	return entity
}

func UpdateCommentReq2Entity(req *types.UpdateCommentReq) *entities.Comment {
	entity := &entities.Comment{}
	entity.Content = req.Content
	entity.Attachments = req.Attachments
	entity.IsTopUp = req.IsTopUp
	return entity
}

func CommentEntities2ResList(entities []*entities.Comment) types.ListCommentRes {
	resList := make([]*types.CommentRes, 0, len(entities))
	for _, e := range entities {
		resList = append(resList, CommentEntity2Res(e))
	}
	return resList
}
