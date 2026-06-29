package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseEventRouter(router *gin.RouterGroup) {
	eventGroup := router.Group(
		"/events",
		middlewares.RateLimiter(64, "events"),
		middlewares.JWTValidator,
	)
	{
		eventGroup.GET("/", controllers.ListEventHandler)
		eventGroup.GET("/:eventId", controllers.GetEventHandler)
		eventGroup.POST("/", controllers.CreateEventHandler)
		eventGroup.PUT("/:eventId", controllers.UpdateEventHandler)
		eventGroup.DELETE("/:eventId", controllers.DeleteEventHandler)
		eventGroup.PUT("/", controllers.BatchUpdateEventHandler)
	}
}
