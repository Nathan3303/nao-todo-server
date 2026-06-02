package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// 检查事项路由组
func UseEventRouter(router *gin.RouterGroup) {
	eventGroup := router.Group(
		"/events",
		middlewares.RateLimiter(64, "events"),
		middlewares.JWTValidator,
	)
	{
		// 获取检查事项列表
		eventGroup.GET("/", controllers.ListEventHandler)

		// 获取检查事项详情
		eventGroup.GET("/:eventId", controllers.GetEventHandler)

		// 创建检查事项
		eventGroup.POST("/", controllers.CreateEventHandler)

		// 更新检查事项
		eventGroup.PUT("/:eventId", controllers.UpdateEventHandler)

		// 删除检查事项
		eventGroup.DELETE("/:eventId", controllers.DeleteEventHandler)

		// 重新排序检查事项

		// 批量更新检查事项
		eventGroup.PUT("/", controllers.BatchUpdateEventHandler)
	}
}
