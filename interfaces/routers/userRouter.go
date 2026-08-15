package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseUserRouter(
	router *gin.RouterGroup,
	ctrl *controllers.UserController,
	sessionCtrl *controllers.SessionController,
	auth authApp.AuthApp,
) {
	userGroup := router.Group(
		"/user",
		middlewares.RateLimiter(auth, 16, "user"),
		middlewares.JWTValidator(auth),
	)
	{
		userGroup.GET("/profile", ctrl.GetUserProfile)
		userGroup.PUT("/nickname", ctrl.UpdateUserNickname)
		userGroup.PUT("/password", ctrl.UpdateUserPassword)
		userGroup.PUT("/avatar", ctrl.UpdateUserAvatar)
		userGroup.DELETE("/", ctrl.DeleteUser)
		userGroup.PUT("/restore", ctrl.RestoreUser)
		userGroup.GET("/config", ctrl.GetUserConfig)
		userGroup.PUT("/config", ctrl.UpdateUserConfig)
		userGroup.GET("/sessions", sessionCtrl.ListUserSessions)
		userGroup.DELETE("/sessions", sessionCtrl.LogoutOtherSessions)
		userGroup.DELETE("/sessions/:sessionId", sessionCtrl.LogoutUserSession)
	}
}
