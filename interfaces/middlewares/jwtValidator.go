package middlewares

import (
	authApp "naotodoserver/application/auth"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func getJwtString(ctx *gin.Context) string {
	// 优先从 query param 获取（兼容 EventSource，不支持自定义 header）
	if token := ctx.Query("token"); token != "" {
		return token
	}
	// 其次从 Authorization header 获取
	jwtRaw := ctx.GetHeader("Authorization")
	jwtString := strings.Split(jwtRaw, "Bearer ")
	jwtString = append(jwtString, "")
	return jwtString[1]
}

func JWTValidator(auth authApp.AuthApp) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 验证 JWT
		jwtString := getJwtString(ctx)
		userId, err := auth.Validate(ctx, jwtString)
		if err != nil {
			// 不向客户端暴露内部错误细节（会话不存在、token 过期等）
			controllers.Failure(ctx, types.ResponseData{
				Code:    10041,
				Message: "用户凭证验证失败",
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
		// 2. 写入用户信息和 token 到上下文
		ctx.Request = ctx.Request.WithContext(
			iCtx.SetToken(
				iCtx.SetUserId(ctx.Request.Context(), int64(userId)),
				jwtString,
			),
		)
		// 3. 检测通过，继续处理请求
		ctx.Next()
	}
}
