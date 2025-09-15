package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateEventHandlerV1DTO struct {
	EventIdRaw  string
	EventId     int64
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      *bool  `json:"isDone"`
	SortId      int32  `json:"sortId"`
}

func UpdateEventHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateEventHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50021,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取检查事项 ID
	if dto.EventIdRaw = ctx.Param("eventId"); dto.EventIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50022,
			Message: "检查事项 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.EventId, _ = strconv.ParseInt(dto.EventIdRaw, 10, 64)

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50023,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 构建更新条件
	var eventCond models.Event
	eventCond.UpdatedAt = time.Time(time.Now())
	if dto.Name != "" {
		eventCond.Name = dto.Name
	}
	if dto.Description != "" {
		eventCond.Description = dto.Description
	}
	if dto.SortId != 0 {
		eventCond.SortId = dto.SortId
	}

	// 更新记录
	result := core.DB.Where("id = ? and user_id = ?", dto.EventId, userId).UpdateColumns(&eventCond)

	// 当 IsDone 存在时强制更新 (避免零值影响)
	if dto.IsDone != nil {
		result.UpdateColumn("is_done", dto.IsDone)
	}

	// 判断更新结果
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50024,
			Message: "检查事项更新失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    50020,
		Message: "检查事项更新成功",
		Data:    eventCond,
	})
}
