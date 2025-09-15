package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateCommentHandlerV1DTO struct {
	CommentIdRaw string
	CommentId    int64
	Content      string   `json:"content"`
	IsTopUp      *bool    `json:"isTopUp"`
	Attachments  []string `json:"attachments"`
}

func UpdateCommentHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateCommentHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60021,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取评论 ID
	if dto.CommentIdRaw = ctx.Param("commentId"); dto.CommentIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60022,
			Message: "评论 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.CommentId, _ = strconv.ParseInt(dto.CommentIdRaw, 10, 64)

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60023,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 构建更新条件
	var commentCond models.Comment
	commentCond.UpdatedAt = time.Time(time.Now())
	if dto.Content != "" {
		commentCond.Content = dto.Content
	}
	if dto.Attachments != nil {
		commentCond.Attachments = dto.Attachments
	}

	// 更新记录
	result := core.DB.Where("id = ? and user_id = ?", dto.CommentId, userId).UpdateColumns(&commentCond)

	// 当 IsTopUp 存在时强制更新 (避免零值影响)
	if dto.IsTopUp != nil {
		result.UpdateColumn("is_top_up", dto.IsTopUp)
	}

	// 判断更新结果
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60024,
			Message: "评论更新失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    60020,
		Message: "评论更新成功",
		Data:    dto.CommentIdRaw,
	})
}
