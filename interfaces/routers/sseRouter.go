package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseSSERouter(router *gin.RouterGroup) {
	sseGroup := router.Group(
		"/sse",
		middlewares.RateLimiter(10, "sse"),
		middlewares.JWTValidator,
	)
	{
		sseGroup.GET("/reminders", controllers.ReminderStream)
	}
}
