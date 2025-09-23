package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"naotodoserver/utils"
	"strconv"
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
		ctx.Abort()
		return
	}

	// 解析 JWT Claims
	iUserJWTClaims, err := utils.ParseUserJWT(dto.JWT)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10042,
			Message: "用户凭证无效",
			Data:    nil,
		})
		ctx.Abort()
		return
	}

	// 转换 UserId
	userId, err := strconv.ParseInt(iUserJWTClaims.Profile.Id, 10, 64)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10044,
			Message: "用户凭证无效",
			Data:    nil,
		})
		ctx.Abort()
		return
	}

	// 验证用户 JWT 是否过期
	if utils.IsUserJWTExpiredByClaims(&iUserJWTClaims) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10043,
			Message: "用户凭证已过期,请重新登录",
			Data:    nil,
		})
		ctx.Abort()
		return
	}

	// 查找是否有对应的 session 记录
	var session models.Session
	var result = core.DB.Where(&models.Session{JWT: dto.JWT, UserId: userId}).First(&session)
	if session.UserId == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10044,
			Message: "用户凭证已过期,请重新登录",
			Data:    nil,
		})
		ctx.Abort()
		return
	}

	// 验证通过，继续处理请求
	ctx.Set("userId", session.UserId)
	ctx.Next()
}
