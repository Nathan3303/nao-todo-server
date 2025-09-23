package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteTodoHandlerV1DTO struct {
	TodoIdRaw       string
	TodoId          int64
	isHardDeleteRaw string
	isHardDelete    bool
}

func DeleteTodoHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto DeleteTodoHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40031,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Param("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40032,
			Message: "待办 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 获取是否硬删除
	dto.isHardDeleteRaw = ctx.Query("hard")
	dto.isHardDelete = dto.isHardDeleteRaw == "true"

	// 执行删除
	var result *gorm.DB
	if dto.isHardDelete {
		result = core.DB.Unscoped().Where("id = ? and user_id = ?", dto.TodoId, userId).Delete(&models.Todo{})
	} else {
		result = core.DB.Where("id = ? and user_id = ?", dto.TodoId, userId).Delete(&models.Todo{})
	}

	// 判断删除结果
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40033,
			Message: "待办删除失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40030,
		Message: "待办删除成功",
		Data:    dto.TodoIdRaw,
	})
}
