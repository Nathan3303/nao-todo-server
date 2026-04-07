package comment

import (
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/valueobjects"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
)

// CommentEntityToRes 评论响应
// @param e 评论实体
// @return 评论响应
func CommentEntityToRes(e *entities.Comment) *types.CommentRes {
	res := &types.CommentRes{}
	res.Id = strconv.FormatInt(e.Id, 10)
	res.TaskId = strconv.FormatInt(e.TaskId, 10)
	res.Content = e.Content
	res.CreatedAt = utils.TimePtr2DateString(&e.CreatedAt)
	res.Attachments = e.Attachments
	res.IsTopUp = e.IsTopUp
	return res
}

// CreateCommentReqToValueObject 创建评论请求体转换为评论值对象
// @param req 创建评论请求体
// @return 评论值对象
func CreateCommentReqToValueObject(
	req *types.CreateCommentReq,
) (*valueobjects.CreateComment, error) {
	taskIdInt64, err := strconv.ParseInt(req.TaskId, 10, 64)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreateComment(
		taskIdInt64,
		req.Content,
		nil,
		false,
	)
}

// UpdateCommentReqToValueObject 更新评论请求体转换为评论值对象
// @param req 更新评论请求体
// @return 评论值对象
func UpdateCommentReqToValueObject(
	req *types.UpdateCommentReq,
) (*valueobjects.UpdateComment, error) {
	return valueobjects.NewUpdateComment(
		req.Content,
		req.Attachments,
		req.IsTopUp,
	)
}

// CommentEntitiesToListRes 评论实体列表转换为评论响应列表
// @param entities 评论实体列表
// @return 评论响应列表
func CommentEntitiesToListRes(entities []*entities.Comment) []*types.CommentRes {
	resList := make([]*types.CommentRes, 0, len(entities))
	for _, e := range entities {
		resList = append(resList, CommentEntityToRes(e))
	}
	return resList
}
