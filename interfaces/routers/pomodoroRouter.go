package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UsePomodoroRouter(router *gin.RouterGroup) {
	pomodoroGroup := router.Group(
		"/pomodoros",
		middlewares.RateLimiter(48, "pomodoros"),
		middlewares.JWTValidator,
	)
	{
		pomodoroGroup.POST("/", controllers.CreatePomodoroHandler)
		pomodoroGroup.GET("/:id", controllers.GetPomodoroHandler)
		pomodoroGroup.GET("/", controllers.ListPomodoroHandler)
	}
}
