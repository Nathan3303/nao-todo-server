package routers

import (
	tagapis "naotodoserver/apis/tag"
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func TagRouterInit(router *gin.RouterGroup) {
	tagRouter := router.Group("/tag", userapis.ValidateHandlerV1)
	{
		tagRouter.GET("/:tagId", tagapis.GetTagHandlerV1)
		tagRouter.POST("/", tagapis.CreateTagHandlerV1)
		tagRouter.PUT("/:tagId", tagapis.UpdateTagHandlerV1)
		tagRouter.DELETE("/:tagId", tagapis.DeleteTagHandlerV1)
		tagRouter.PUT("/restore/:tagId", tagapis.RestoreTagHandlerV1)
		tagRouter.PUT("/preference/:tagId", tagapis.UpdateTagPreferenceHandlerV1)
	}
	tagsRouter := router.Group("/tags", userapis.ValidateHandlerV1)
	{
		tagsRouter.GET("/", tagapis.GetTagsHandlerV1)
	}
}
