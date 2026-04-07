package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

/* 用户路由 */
func UseUserRouter(router *gin.RouterGroup) {
	userGroup := router.Group(
		"/user",
		middlewares.RateLimiter(10, "user"),
		middlewares.JWTValidator,
	)
	{
		/* 获取用户信息路由 */
		userGroup.GET("/profile", controllers.GetUserProfileHandler)

		/* 更新用户信息路由 */
		userGroup.PUT("/nickname", controllers.UpdateUserNicknameHandler)

		/* 更新用户密码路由 */
		userGroup.PUT("/password", controllers.UpdateUserPasswordHandler)

		/* 更新用户头像路由 */
		userGroup.PUT("/avatar", controllers.UpdateUserAvatarHandler)

		/* 激活用户路由 */
		userGroup.PUT("/active", controllers.ActiveUserHandler)

		/* 禁用用户路由 */
		userGroup.PUT("/deactive", controllers.DeactiveUserHandler)
	}
}
