package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseTagRouter(router *gin.RouterGroup, ctrl *controllers.TagController) {
	tagGroup := router.Group(
		"/tags",
		middlewares.RateLimiter(32, "tags"),
		middlewares.JWTValidator,
	)
	{
		tagGroup.GET("/", ctrl.ListTag)
		tagGroup.GET("/:tagId", ctrl.GetTag)
		tagGroup.POST("/", ctrl.CreateTag)
		tagGroup.PUT("/", ctrl.BatchUpdateTags)
		tagGroup.PUT("/:tagId", ctrl.UpdateTag)
		tagGroup.DELETE("/:tagId", ctrl.DeleteTag)
		tagGroup.GET(
			"/:tagId/preference",
			ctrl.GetTagPreference,
		)
		tagGroup.POST(
			"/:tagId/preference",
			ctrl.UpdateTagPreference,
		)
	}
}
