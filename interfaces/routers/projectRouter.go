package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseProjectRouter(router *gin.RouterGroup) {
	projectGroup := router.Group(
		"/projects",
		middlewares.RateLimiter(32, "projects"),
		middlewares.JWTValidator,
	)
	{
		projectGroup.GET(
			"/",
			controllers.ListProjectHandler,
		)
		projectGroup.GET(
			"/:projectId",
			controllers.GetProjectHandler,
		)
		projectGroup.POST(
			"/",
			controllers.CreateProjectHandler,
		)
		projectGroup.PUT(
			"/",
			controllers.BatchUpdateProjectsHandler,
		)
		projectGroup.PUT(
			"/:projectId",
			controllers.UpdateProjectHandler,
		)
		projectGroup.DELETE(
			"/:projectId",
			controllers.DeleteProjectHandler,
		)
		projectGroup.PUT(
			"/restore/:projectId",
			controllers.RestoreProjectHandler,
		)
		projectGroup.PUT(
			"/archive/:projectId",
			controllers.ArchiveProjectHandler,
		)
		projectGroup.PUT(
			"/unarchive/:projectId",
			controllers.UnarchiveProjectHandler,
		)
		projectGroup.GET(
			"/:projectId/preference",
			controllers.GetProjectPreferenceHandler,
		)
		projectGroup.POST(
			"/:projectId/preference",
			controllers.SaveProjectPreferenceHandler,
		)
	}
}
