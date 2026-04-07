package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

/* 标签路由 */
func UseTagRouter(router *gin.RouterGroup) {
	tagGroup := router.Group(
		"/tags",
		middlewares.RateLimiter(20, "tags"),
		middlewares.JWTValidator,
	)
	{
		/* 获取标签列表路由 */
		tagGroup.GET("/", controllers.ListTagHandler)

		/* 获取标签详情路由 */
		tagGroup.GET("/:tagId", controllers.GetTagHandler)

		/* 创建标签路由 */
		tagGroup.POST("/", controllers.CreateTagHandler)

		/* 更新标签路由 */
		tagGroup.PUT("/:tagId", controllers.UpdateTagHandler)

		/* 删除标签路由 */
		tagGroup.DELETE("/:tagId", controllers.DeleteTagHandler)

		/* 获取标签偏好路由 */
		tagGroup.GET(
			"/:tagId/preference",
			controllers.GetTagPreferenceHandler,
		)

		/* 更新标签偏好路由 */
		tagGroup.POST(
			"/:tagId/preference",
			controllers.UpdateTagPreferenceHandler,
		)
	}
}
