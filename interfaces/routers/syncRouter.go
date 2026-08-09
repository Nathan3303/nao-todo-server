package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UseSyncRouter 配置数据同步路由
func UseSyncRouter(
	router *gin.RouterGroup,
	ctrl *controllers.SyncController,
	auth authApp.AuthApp,
) {
	syncGroup := router.Group(
		"/sync",
		middlewares.RateLimiter(auth, 24, "sync"),
		middlewares.JWTValidator(auth),
	)
	{
		syncGroup.POST("/push", ctrl.Push)
		syncGroup.POST("/pull", ctrl.Pull)
	}
}
