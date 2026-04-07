package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

/* 项目路由 */
func UseProjectRouter(router *gin.RouterGroup) {
	projectGroup := router.Group(
		"/projects",
		middlewares.RateLimiter(20, "projects"),
		middlewares.JWTValidator,
	)
	{
		/* 获取项目列表路由 */
		projectGroup.GET(
			"/",
			controllers.ListProjectHandler,
		)

		/* 获取项目详情路由 */
		projectGroup.GET(
			"/:projectId",
			controllers.GetProjectHandler,
		)

		/* 创建项目路由 */
		projectGroup.POST(
			"/",
			controllers.CreateProjectHandler,
		)

		/* 更新项目路由 */
		projectGroup.PUT(
			"/:projectId",
			controllers.UpdateProjectHandler,
		)

		/* 删除项目路由 */
		projectGroup.DELETE(
			"/:projectId",
			controllers.DeleteProjectHandler,
		)

		/* 恢复项目路由 */
		projectGroup.PUT(
			"/restore/:projectId",
			controllers.RestoreProjectHandler,
		)

		/* 归档项目路由 */
		projectGroup.PUT(
			"/archive/:projectId",
			controllers.ArchiveProjectHandler,
		)

		/* 取消归档项目路由 */
		projectGroup.PUT(
			"/unarchive/:projectId",
			controllers.UnarchiveProjectHandler,
		)

		/* 获取项目偏好路由 */
		projectGroup.GET(
			"/:projectId/preference",
			controllers.GetProjectPreferenceHandler,
		)

		/* 保存项目偏好路由 */
		projectGroup.POST(
			"/:projectId/preference",
			controllers.SaveProjectPreferenceHandler,
		)
	}
}
