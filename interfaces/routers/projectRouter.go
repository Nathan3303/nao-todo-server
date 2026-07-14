package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseProjectRouter(router *gin.RouterGroup, ctrl *controllers.ProjectController, auth authApp.AuthApp) {
	projectGroup := router.Group(
		"/projects",
		middlewares.RateLimiter(auth, 32, "projects"),
		middlewares.JWTValidator(auth),
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
