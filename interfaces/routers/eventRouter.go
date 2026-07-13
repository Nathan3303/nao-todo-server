package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseEventRouter(router *gin.RouterGroup, ctrl *controllers.EventController) {
	eventGroup := router.Group(
		"/events",
		middlewares.RateLimiter(64, "events"),
		middlewares.JWTValidator,
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
