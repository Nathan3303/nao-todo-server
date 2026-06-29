package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseUserRouter(router *gin.RouterGroup) {
	userGroup := router.Group(
		"/user",
		middlewares.RateLimiter(16, "user"),
		middlewares.JWTValidator,
	)
	{
		userGroup.GET("/profile", controllers.GetUserProfileHandler)
		userGroup.PUT("/nickname", controllers.UpdateUserNicknameHandler)
		userGroup.PUT("/password", controllers.UpdateUserPasswordHandler)
		userGroup.PUT("/avatar", controllers.UpdateUserAvatarHandler)
		userGroup.PUT("/active", controllers.ActiveUserHandler)
		userGroup.PUT("/deactive", controllers.DeactiveUserHandler)
		userGroup.GET("/config", controllers.GetUserConfigHandler)
		userGroup.PUT("/config", controllers.UpdateUserConfigHandler)
	}
}
