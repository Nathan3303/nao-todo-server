package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RestoreTodoHandlerV1DTO struct {
	TodoId    int64
	TodoIdRaw string
}

func RestoreTodoHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto RestoreTodoHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40041,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Param("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40042,
			Message: "待办 ID 不能为空",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 恢复记录
	var todo models.Todo
	core.DB.Unscoped().Where("id = ? and user_id = ?", dto.TodoId, userId).First(&todo)
	if todo.DeletedAt.Valid { // 确认已被软删除
		todo.DeletedAt = gorm.DeletedAt{} // 手动清空 DeletedAt
		core.DB.Unscoped().Save(&todo)    // 保存，恢复记录
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40040,
		Message: "待办恢复成功",
		Data:    todo.ID,
	})
}
