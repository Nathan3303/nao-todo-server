package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"naotodoserver/utils"

	"github.com/gin-gonic/gin"
)

type SignOutHandlerV1DTO struct {
	JWT string
}

func SignOutHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto SignOutHandlerV1DTO

	// 获取参数
	if dto.JWT = ctx.Query("jwt"); dto.JWT == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10031,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 验证用户 JWT 是否过期
	iUserJWTClaims, err := utils.ParseUserJWT(dto.JWT)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10032,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}
	if utils.IsUserJWTExpiredByClaims(&iUserJWTClaims) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10033,
			Message: "用户凭证已过期",
			Data:    nil,
		})
		return
	}

	// 查找并删除数据库对应的 session 记录
	result := core.DB.Where(&models.Session{JWT: dto.JWT}).Delete(&models.Session{})
	if result.Error != nil {
		apis.Success(ctx, apis.ResponseData{
			Code:    10034,
			Message: "登出成功",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    10030,
		Message: "登出成功",
		Data:    nil,
	})
}
