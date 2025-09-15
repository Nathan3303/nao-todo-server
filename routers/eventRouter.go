package routers

import (
	eventapis "naotodoserver/apis/event"
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func EventRouterInit(router *gin.RouterGroup) {
	eventRouter := router.Group("/event", userapis.ValidateHandlerV1)
	{
		eventRouter.POST("/", eventapis.CreateEventHandlerV1)
		eventRouter.PUT("/:eventId", eventapis.UpdateEventHandlerV1)
		eventRouter.DELETE("/:eventId", eventapis.DeleteEventHandlerV1)
	}
	eventsRouter := router.Group("/events", userapis.ValidateHandlerV1)
	{
		eventsRouter.GET("/", eventapis.GetEventsHandlerV1)
	}
}
