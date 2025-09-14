package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"
	"naotodoserver/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type ValidateHandlerV1DTO struct {
	JWT string
}

func ValidateHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto ValidateHandlerV1DTO

	// 获取参数
	splited := strings.Split(ctx.GetHeader("Authorization"), "Bearer ")
	splited = append(splited, "")
	if dto.JWT = splited[1]; dto.JWT == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10041,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 验证用户 JWT 是否过期
	iUserJWTClaims, err := utils.ParseUserJWT(dto.JWT)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10042,
			Message: "用户凭证无效",
			Data:    dto.JWT,
		})
		return
	}
	if utils.IsUserJWTExpiredByClaims(&iUserJWTClaims) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10043,
			Message: "用户凭证已过期,请重新登录",
			Data:    dto.JWT,
		})
		return
	}

	// 查找是否有对应的 session 记录
	var session modules.Session
	core.DB.Where(&modules.Session{JWT: dto.JWT, UserId: iUserJWTClaims.Profile.Id}).First(&session)
	if session.UserId == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10044,
			Message: "用户凭证已过期,请重新登录",
			Data:    "",
		})
		return
	}

	// 返回数据
	// apis.Success(ctx, apis.ResponseData{
	// 	Code:    10040,
	// 	Message: "用户凭证验证通过",
	// 	Data:    nil,
	// })

	// 验证通过，继续处理请求
	ctx.Set("userId", session.UserId)
	ctx.Next()
}
