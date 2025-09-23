package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"

	"github.com/gin-gonic/gin"
)

type UpdateProfileHandlerV1DTO struct {
	Nickname string `json:"nickname"`
}

func UpdateProfileHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateProfileHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10051,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10052,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 构建更新条件
	var userCond models.User
	if dto.Nickname != "" {
		userCond.Nickname = dto.Nickname
	}

	// 更新用户信息
	result := core.DB.Where("id = ?", userId).UpdateColumns(&userCond)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10053,
			Message: "用户信息更新失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    10050,
		Message: "用户信息更新成功",
		Data:    nil,
	})
}
