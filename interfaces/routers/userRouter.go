package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseUserRouter(router *gin.RouterGroup, ctrl *controllers.UserController) {
	userGroup := router.Group(
		"/user",
		middlewares.RateLimiter(16, "user"),
		middlewares.JWTValidator,
	)
	{
		userGroup.GET("/profile", ctrl.GetUserProfile)
		userGroup.PUT("/nickname", ctrl.UpdateUserNickname)
		userGroup.PUT("/password", ctrl.UpdateUserPassword)
		userGroup.PUT("/avatar", ctrl.UpdateUserAvatar)
		userGroup.PUT("/active", ctrl.ActiveUser)
		userGroup.PUT("/deactive", ctrl.DeactiveUser)
		userGroup.GET("/config", ctrl.GetUserConfig)
		userGroup.PUT("/config", ctrl.UpdateUserConfig)
	}
}
