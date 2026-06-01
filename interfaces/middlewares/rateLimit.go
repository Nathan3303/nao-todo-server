package middlewares

import (
	"naotodoserver/application"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"

	"net/http"

	"github.com/gin-gonic/gin"
)

func RateLimiter(limit int8, tag string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 构建 Key
		key := ctx.ClientIP() + ":" + tag
		// 2. 检查
		err := application.App.Auth.RateLimit(ctx, key, limit)
		if err != nil {
			controllers.FailureByHttpStatus(ctx, http.StatusTooManyRequests, types.ResponseData{
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
