package comment

import (
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/valueobjects"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
)

// CommentEntityToRes 评论实体转换为评论响应
func CommentEntityToRes(e *entities.Comment) *types.CommentRes {
	return &types.CommentRes{
		Id:          strconv.FormatInt(e.Id, 10),
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		Content:     e.Content,
		Attachments: e.Attachments,
		IsTopUp:     e.IsTopUp,
		Nickname:    e.Nickname,
		Avatar:      e.Avatar,
		CreatedAt:   utils.TimePtr2DateString(&e.CreatedAt),
		UpdatedAt:   utils.TimePtr2DateString(&e.UpdatedAt),
	}
}

// CreateCommentReqToValueObject 创建评论请求体转换为评论值对象
// @param userId 用户 ID
// @param req 创建评论请求体
// @return 评论值对象
func CreateCommentReqToValueObject(
	userId int64,
	req *types.CreateCommentReq,
) (*valueobjects.CreateComment, error) {
	taskIdInt64, err := strconv.ParseInt(req.TaskId, 10, 64)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreateComment(
		userId,
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
