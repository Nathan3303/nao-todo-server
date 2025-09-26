package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteCommentHandlerV1DTO struct {
	CommentIdRaw string
	CommentId    int64
}

func DeleteCommentHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto DeleteCommentHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60031,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取评论 ID
	if dto.CommentIdRaw = ctx.Param("commentId"); dto.CommentIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60032,
			Message: "评论 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.CommentId, _ = strconv.ParseInt(dto.CommentIdRaw, 10, 64)

	// 执行删除
	result := core.DB.Where("id = ? and user_id = ?", dto.CommentId, userId).Delete(&models.Comment{})
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60033,
			Message: "评论删除失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    60030,
		Message: "评论删除成功",
		Data:    dto.CommentIdRaw,
	})
}
