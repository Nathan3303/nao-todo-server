package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// SSE 路由组
func UseSSERouter(router *gin.RouterGroup) {
	sseGroup := router.Group(
		"/sse",
		middlewares.RateLimiter(10, "sse"),
		middlewares.JWTValidator,
	)
	{
		// 提醒事件 SSE 流
		sseGroup.GET("/reminders", controllers.ReminderStream)
	}
}
