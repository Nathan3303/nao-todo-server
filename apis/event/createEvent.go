package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateEventHandlerV1DTO struct {
	TodoIdRaw   string `json:"todoId" binding:"required"`
	TodoId      int64
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func CreateEventHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto CreateEventHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50011,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50012,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	var err error
	dto.TodoId, err = strconv.ParseInt(dto.TodoIdRaw, 10, 64)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50013,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 创建记录
	var event = &models.Event{
		UserId:      userId.(int64),
		TodoId:      dto.TodoId,
		Name:        dto.Name,
		Description: dto.Description,
		IsDone:      false,
		SortId:      int32(time.Now().Unix()),
	}
	result := core.DB.Create(&event)
	if event.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50014,
			Message: "检查事项创建失败",
			Data:    nil,
		})
		return
	}

	// 返回成功结果
	apis.Success(ctx, apis.ResponseData{
		Code:    50010,
		Message: "检查事项创建成功",
		Data:    ToEventResponse(event),
	})
}
