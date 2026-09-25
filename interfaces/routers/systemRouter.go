package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UseSystemRouter 配置系统路由
func UseSystemRouter(
	router *gin.RouterGroup,
	ctrl *controllers.SystemController,
	auth authApp.AuthApp,
) {
	systemGroup := router.Group(
		"/system",
		middlewares.RateLimiter(auth, 48, "system"),
		middlewares.JWTValidator(auth),
	)
	{
		systemGroup.GET("/config", ctrl.GetSystemConfig)
	}
}
