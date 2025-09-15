package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetEventsHandlerV1DTO struct {
	TodoIdRaw string
	TodoId    int64
}

func GetEventsHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetEventsHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50001,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Query("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50002,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 获取待办事项列表
	var events []models.Event
	result := core.DB.Where("todo_id = ? and user_id = ?", dto.TodoId, userId).Order("sort_id ASC").Find(&events)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    50003,
			Message: "待办事项列表查询失败",
			Data:    nil,
		})
		return
	}

	// 返回待办事项列表信息
	apis.Success(ctx, apis.ResponseData{
		Code:    50000,
		Message: "获取待办事项列表成功",
		Data:    events,
	})
}
