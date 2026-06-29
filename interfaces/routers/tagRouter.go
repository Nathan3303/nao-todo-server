package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseTagRouter(router *gin.RouterGroup) {
	tagGroup := router.Group(
		"/tags",
		middlewares.RateLimiter(32, "tags"),
		middlewares.JWTValidator,
	)
	{
		tagGroup.GET("/", controllers.ListTagHandler)
		tagGroup.GET("/:tagId", controllers.GetTagHandler)
		tagGroup.POST("/", controllers.CreateTagHandler)
		tagGroup.PUT("/", controllers.BatchUpdateTagsHandler)
		tagGroup.PUT("/:tagId", controllers.UpdateTagHandler)
		tagGroup.DELETE("/:tagId", controllers.DeleteTagHandler)
		tagGroup.GET(
			"/:tagId/preference",
			controllers.GetTagPreferenceHandler,
		)
		tagGroup.POST(
			"/:tagId/preference",
			controllers.UpdateTagPreferenceHandler,
		)
	}
}
