package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseProjectRouter(router *gin.RouterGroup, ctrl *controllers.ProjectController) {
	projectGroup := router.Group(
		"/projects",
		middlewares.RateLimiter(32, "projects"),
		middlewares.JWTValidator,
	)
	{
		projectGroup.GET(
			"/",
			ctrl.ListProject,
		)
		projectGroup.GET(
			"/:projectId",
			ctrl.GetProject,
		)
		projectGroup.POST(
			"/",
			ctrl.CreateProject,
		)
		projectGroup.PUT(
			"/",
			ctrl.BatchUpdateProjects,
		)
		projectGroup.PUT(
			"/:projectId",
			ctrl.UpdateProject,
		)
		projectGroup.DELETE(
			"/:projectId",
			ctrl.DeleteProject,
		)
		projectGroup.PUT(
			"/restore/:projectId",
			ctrl.RestoreProject,
		)
		projectGroup.PUT(
			"/archive/:projectId",
			ctrl.ArchiveProject,
		)
		projectGroup.PUT(
			"/unarchive/:projectId",
			ctrl.UnarchiveProject,
		)
		projectGroup.GET(
			"/:projectId/preference",
			ctrl.GetProjectPreference,
		)
		projectGroup.POST(
			"/:projectId/preference",
			ctrl.SaveProjectPreference,
		)
	}
}
