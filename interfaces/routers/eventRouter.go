package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseEventRouter(router *gin.RouterGroup, ctrl *controllers.EventController, auth authApp.AuthApp) {
	eventGroup := router.Group(
		"/events",
		middlewares.RateLimiter(auth, 64, "events"),
		middlewares.JWTValidator(auth),
	)
	{
		eventGroup.GET("/", ctrl.ListEvent)
		eventGroup.GET("/:eventId", ctrl.GetEvent)
		eventGroup.POST("/", ctrl.CreateEvent)
		eventGroup.PUT("/:eventId", ctrl.UpdateEvent)
		eventGroup.DELETE("/:eventId", ctrl.DeleteEvent)
		eventGroup.PUT("/", ctrl.BatchUpdateEvent)
	}
}
