package middlewares

import (
	"naotodoserver/application/auth"
	"naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func getJwtString(ctx *gin.Context) string {
	jwtRaw := ctx.GetHeader("Authorization")
	jwtString := strings.Split(jwtRaw, "Bearer ")
	jwtString = append(jwtString, "")
	return jwtString[1]
}

func JWTValidator(ctx *gin.Context) {
	// 1. 验证 JWT
	jwtString := getJwtString(ctx)
	userId, err := auth.App.Validate(ctx, jwtString)
	if err != nil {
		controllers.Failure(ctx, types.ResponseData{
			Code:    10041,
			Message: "用户凭证验证失败",
			Data:    err.Error(),
		})
		ctx.Abort()
		return
	}
	if userId <= 0 {
		controllers.Failure(ctx, types.ResponseData{
			Code:    10042,
			Message: "用户凭证验证失败",
		})
		ctx.Abort()
		return
	}
	// 2. 写入用户信息到上下文
	ctx.Request = ctx.Request.WithContext(
		context.SetUserId(ctx.Request.Context(), userId),
	)
	// 3. 检测通过，继续处理请求
	ctx.Next()
}
