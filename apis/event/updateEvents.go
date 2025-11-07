package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type updateOptions struct {
	EventIdRaw  string `json:"eventId" binding:"required"`
	EventId     int64
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      *bool  `json:"isDone"`
	SortId      int32  `json:"sortId"`
}

type UpdateEventsHandlerV1DTO struct {
	Updates []updateOptions `json:"updates" form:"updates" binding:"required"`
}

func UpdateEventsHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateEventsHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50041,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50042,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 更新事务
	err := core.DB.Transaction(func(tx *gorm.DB) error {
		for _, update := range dto.Updates {
			updateEvent := models.Event{}
			update.EventId, _ = strconv.ParseInt(update.EventIdRaw, 10, 64)
			if update.Name != "" {
				updateEvent.Name = update.Name
			}
			if update.Description != "" {
				updateEvent.Description = update.Description
			}
			if update.IsDone == nil {
				updateEvent.IsDone = false
			} else {
				updateEvent.IsDone = *update.IsDone
			}
			if update.SortId != 0 {
				updateEvent.SortId = update.SortId
			}
			result := core.DB.Model(&models.Event{}).Where("user_id = ? and id = ?", userId, update.EventId).UpdateColumns(&updateEvent)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})

	// 处理失败结果
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50043,
			Message: "检查事项更新失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    50040,
		Message: "检查事项更新成功",
		Data:    dto.Updates,
	})

}
