package routers

import (
	"naotodoserver/apis"
	projectapis "naotodoserver/apis/project"
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func defaultHandler(ctx *gin.Context) {
	apis.Success(ctx, apis.ResponseData{
		Code:    200,
		Message: "success",
		Data:    nil,
	})
}

func ProjectRouterInit(router *gin.RouterGroup) {
	projectRouter := router.Group("/project", userapis.ValidateHandlerV1)
	{
		projectRouter.GET("/", projectapis.GetProjectHandlerV1)
		projectRouter.POST("/", projectapis.CreateProjectHandlerV1)
		projectRouter.PUT("/:projectId", projectapis.UpdateProjectHandlerV1)
	}
	projectsRouter := router.Group("/projects")
	{
		projectsRouter.GET("/", defaultHandler)

	}
}
