package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetCommentsHandlerV1DTO struct {
	TodoIdRaw string
	TodoId    int64
}

func GetCommentsHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetCommentsHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60001,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Query("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60002,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 获取评论列表
	var comments []models.Comment
	result := core.DB.Preload("CommentUser").Where("todo_id = ? and user_id = ?", dto.TodoId, userId).Find(&comments)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60003,
			Message: "评论列表查询失败",
			Data:    nil,
		})
		return
	}

	// 返回评论列表信息
	apis.Success(ctx, apis.ResponseData{
		Code:    60000,
		Message: "获取评论列表成功",
		Data:    ToCommentResponseList(comments),
	})
}
