package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseTaskRouter(router *gin.RouterGroup, ctrl *controllers.TaskController) {
	taskGroup := router.Group(
		"/tasks",
		middlewares.RateLimiter(48, "tasks"),
		middlewares.JWTValidator,
	)
	{
		taskGroup.GET("/", ctrl.ListTask)
		taskGroup.GET("/:taskId", ctrl.GetTask)
		taskGroup.POST("/", ctrl.CreateTask)
		taskGroup.PUT("/:taskId", ctrl.UpdateTask)
		taskGroup.DELETE("/:taskId", ctrl.DeleteTask)
		taskGroup.PUT("/restore/:taskId", ctrl.RestoreTask)
		taskGroup.POST("/copy/:taskId", ctrl.CopyTask)
		taskGroup.POST("/:taskId/snooze", ctrl.SnoozeTask)
	}
}
