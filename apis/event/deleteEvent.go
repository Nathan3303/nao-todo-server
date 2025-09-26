package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteEventHandlerV1DTO struct {
	EventIdRaw string
	EventId    int64
}

func DeleteEventHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto DeleteEventHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50031,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取检查事项 ID
	if dto.EventIdRaw = ctx.Param("eventId"); dto.EventIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50032,
			Message: "检查事项 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.EventId, _ = strconv.ParseInt(dto.EventIdRaw, 10, 64)

	// 执行删除
	result := core.DB.Where("id = ? and user_id = ?", dto.EventId, userId).Delete(&models.Event{})
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50033,
			Message: "检查事项删除失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    50030,
		Message: "检查事项删除成功",
		Data:    dto.EventIdRaw,
	})
}
