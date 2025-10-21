package routers

import (
	projectapis "naotodoserver/apis/project"
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func ProjectRouterInit(router *gin.RouterGroup) {
	projectRouter := router.Group("/project", userapis.ValidateHandlerV1)
	{
		projectRouter.GET("/:projectId", projectapis.GetProjectHandlerV1)
		projectRouter.POST("/", projectapis.CreateProjectHandlerV1)
		projectRouter.PUT("/:projectId", projectapis.UpdateProjectHandlerV1)
		projectRouter.DELETE("/:projectId", projectapis.DeleteProjectHandlerV1)
		projectRouter.PUT("/restore/:projectId", projectapis.RestoreProjectHandlerV1)
		projectRouter.PUT("/preference/:projectId", projectapis.UpdateProjectPreferenceHandlerV1)
	}
	projectsRouter := router.Group("/projects", userapis.ValidateHandlerV1)
	{
		projectsRouter.GET("/", projectapis.GetProjectsHandlerV1)
	}
}
