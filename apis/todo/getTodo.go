package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetTodoHandlerV1DTO struct {
	TodoIdRaw string `json:"todoId" binding:"required"`
	TodoId    int64
}

func GetTodoHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetTodoHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40001,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Param("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40002,
			Message: "待办 ID 不能为空",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 获取待办
	var todoRaw models.Todo
	result := core.DB.Where("id = ? and user_id = ?", dto.TodoId, userId).First(&todoRaw)
	if todoRaw.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40003,
			Message: "获取待办失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 转换为响应数据并返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40000,
		Message: "获取待办成功",
		Data:    ToTodoResponse(&todoRaw),
	})
}
