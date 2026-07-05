package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UsePomodoroRouter 配置 Pomodoro 路由
func UsePomodoroRouter(router *gin.RouterGroup) {
	// Pomodoro 路由
	pomodoroGroup := router.Group(
		"/pomodoros",
		middlewares.RateLimiter(48, "pomodoros"),
		middlewares.JWTValidator,
	)
	{
		pomodoroGroup.GET("/:id", controllers.GetPomodoroHandler)
		pomodoroGroup.POST("/", controllers.CreatePomodoroHandler)
		pomodoroGroup.PUT("/:id", controllers.UpdatePomodoroHandler)
		pomodoroGroup.DELETE("/:id", controllers.DeletePomodoroHandler)
		pomodoroGroup.PUT("/:id/archived", controllers.ArchivedPomodoroHandler)
		pomodoroGroup.PUT("/:id/unarchived", controllers.UnarchivedPomodoroHandler)
		pomodoroGroup.GET("/", controllers.ListPomodoroHandler)
	}

	// Pomodoro Record 路由
	pomodoroRecordsGroup := router.Group(
		"/pomodoro-records",
		middlewares.RateLimiter(48, "pomodoro-records"),
		middlewares.JWTValidator,
	)
	{
		pomodoroRecordsGroup.POST(
			"/",
			controllers.CreatePomodoroRecordHandler,
		)
		pomodoroRecordsGroup.GET(
			"/:id",
			controllers.GetPomodoroRecordHandler,
		)
		pomodoroRecordsGroup.GET(
			"/",
			controllers.ListPomodoroRecordHandler,
		)
	}
}
