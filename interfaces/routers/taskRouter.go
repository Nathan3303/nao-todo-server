package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// taskRateLimits tasks 路由组桶配额表（固定窗口 60s / IP / 每分钟）：调整配额只改此处。
// - tasks：详情与全部写路径（创建/更新/删除/恢复/复制/稍后提醒）
// - tasks-read：列表只读批量拉取（SEA-01，搜索/日历等全量分页枚举不受详情写配额拖累）
const (
	rateLimitTasks     int64 = 48
	rateLimitTasksRead int64 = 400
)

func UseTaskRouter(
	router *gin.RouterGroup,
	ctrl *controllers.TaskController,
	auth authApp.AuthApp,
) {
	// 列表只读桶（SEA-01-REL-1）：批量列表形态独立配额，不与详情/写共享 48 桶
	readGroup := router.Group(
		"/tasks",
		middlewares.RateLimiter(auth, rateLimitTasksRead, "tasks-read"),
		middlewares.JWTValidator(auth),
	)
	{
		readGroup.GET("/", ctrl.ListTask)
	}
	// 详情与写路径桶（原 tasks 桶语义不变）
	taskGroup := router.Group(
		"/tasks",
		middlewares.RateLimiter(auth, rateLimitTasks, "tasks"),
		middlewares.JWTValidator(auth),
	)
	{
		taskGroup.GET("/:taskId", ctrl.GetTask)
		taskGroup.POST("/", ctrl.CreateTask)
		taskGroup.PUT("/:taskId", ctrl.UpdateTask)
		taskGroup.DELETE("/:taskId", ctrl.DeleteTask)
		taskGroup.PUT("/restore/:taskId", ctrl.RestoreTask)
		taskGroup.POST("/copy/:taskId", ctrl.CopyTask)
		taskGroup.POST("/:taskId/snooze", ctrl.SnoozeTask)
	}
}
