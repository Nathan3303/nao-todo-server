package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UsePomodoroRouter 配置 Pomodoro 路由
func UsePomodoroRouter(router *gin.RouterGroup, ctrl *controllers.PomodoroController) {
	// Pomodoro 路由
	pomodoroGroup := router.Group(
		"/pomodoros",
		middlewares.RateLimiter(48, "pomodoros"),
		middlewares.JWTValidator,
	)
	{
		pomodoroGroup.GET("/:id", ctrl.GetPomodoro)
		pomodoroGroup.POST("/", ctrl.CreatePomodoro)
		pomodoroGroup.PUT("/:id", ctrl.UpdatePomodoro)
		pomodoroGroup.DELETE("/:id", ctrl.DeletePomodoro)
		pomodoroGroup.PUT("/:id/archived", ctrl.ArchivedPomodoro)
		pomodoroGroup.PUT("/:id/unarchived", ctrl.UnarchivedPomodoro)
		pomodoroGroup.GET("/", ctrl.ListPomodoro)
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
			ctrl.CreatePomodoroRecord,
		)
		pomodoroRecordsGroup.GET(
			"/:id",
			ctrl.GetPomodoroRecord,
		)
		pomodoroRecordsGroup.GET(
			"/",
			ctrl.ListPomodoroRecord,
		)
	}
}
