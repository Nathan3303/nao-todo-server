package middlewares

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

func RateLimiter(auth authApp.AuthApp, limit int64, tag string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 构建 Key
		key := ctx.ClientIP() + ":" + tag
		// 2. 检查
		err := auth.RateLimit(ctx, key, limit)
		if err != nil {
			controllers.Failure(ctx, types.ResponseData{
				Code:    10051,
				Message: "请求失败",
				Error:   "请求次数超出单位时间内次数",
			})
			ctx.Abort()
			return
		}
		// 3. next
		ctx.Next()
	}
}
