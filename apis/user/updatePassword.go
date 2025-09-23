package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"

	"github.com/gin-gonic/gin"
)

type UpdatePasswordHandlerV1DTO struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func UpdatePasswordHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdatePasswordHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10061,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10062,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 查询旧密码是否正确
	var user models.User
	result := core.DB.Model(&models.User{}).Where("id = ? and password = ?", userId, dto.OldPassword).First(&user)
	if user.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10063,
			Message: "密码错误",
			Data:    nil,
		})
		return
	}

	// 更新密码
	result = core.DB.Model(&models.User{}).Where("id = ? and password = ?", userId, dto.OldPassword).UpdateColumn("password", dto.NewPassword)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10064,
			Message: "密码更新失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 删除所有会话记录
	result = core.DB.Model(&models.Session{}).Where("user_id = ?", userId).Delete(&models.Session{})
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10065,
			Message: "会话记录删除出错",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    10060,
		Message: "密码更新成功",
		Data:    userId,
	})
}
