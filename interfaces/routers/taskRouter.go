package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// 待办任务路由组
func UseTaskRouter(router *gin.RouterGroup) {
	taskGroup := router.Group(
		"/tasks",
		middlewares.RateLimiter(48, "tasks"),
		middlewares.JWTValidator,
	)
	{
		// 获取待办任务列表路由
		taskGroup.GET("/", controllers.ListTaskHandler)

		// 获取待办任务详情路由
		taskGroup.GET("/:taskId", controllers.GetTaskHandler)

		// 创建待办任务路由
		taskGroup.POST("/", controllers.CreateTaskHandler)

		// 更新待办任务路由
		taskGroup.PUT("/:taskId", controllers.UpdateTaskHandler)

		// 删除待办任务路由
		taskGroup.DELETE("/:taskId", controllers.DeleteTaskHandler)

		// 恢复待办任务路由
		taskGroup.PUT("/restore/:taskId", controllers.RestoreTaskHandler)

		// 复制待办任务路由
		taskGroup.POST("/copy/:taskId", controllers.CopyTaskHandler)
	}
}
