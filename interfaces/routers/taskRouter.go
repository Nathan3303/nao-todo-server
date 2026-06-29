package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseTaskRouter(router *gin.RouterGroup) {
	taskGroup := router.Group(
		"/tasks",
		middlewares.RateLimiter(48, "tasks"),
		middlewares.JWTValidator,
	)
	{
		taskGroup.GET("/", controllers.ListTaskHandler)
		taskGroup.GET("/:taskId", controllers.GetTaskHandler)
		taskGroup.POST("/", controllers.CreateTaskHandler)
		taskGroup.PUT("/:taskId", controllers.UpdateTaskHandler)
		taskGroup.DELETE("/:taskId", controllers.DeleteTaskHandler)
		taskGroup.PUT("/restore/:taskId", controllers.RestoreTaskHandler)
		taskGroup.POST("/copy/:taskId", controllers.CopyTaskHandler)
		taskGroup.POST("/:taskId/snooze", controllers.SnoozeTaskHandler)
	}
}
