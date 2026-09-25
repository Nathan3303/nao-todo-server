package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseSSERouter(router *gin.RouterGroup, ctrl *controllers.SSEController, auth authApp.AuthApp) {
	sseGroup := router.Group(
		"/sse",
		middlewares.RateLimiter(auth, 10, "sse"),
		middlewares.JWTValidator(auth),
	)
	{
		sseGroup.GET("/reminders", ctrl.ReminderStream)
	}
}
