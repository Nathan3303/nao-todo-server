package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

/* 认证路由 */
func UseAuthRouter(router *gin.RouterGroup) {
	authGroup := router.Group(
		"/auth",
		middlewares.RateLimiter(10, "auth"),
	)
	{
		/* 登录路由 */
		authGroup.POST("/signin", controllers.UserSignInHandler)

		/* 注册路由 */
		authGroup.POST("/signup", controllers.UserSignUpHandler)

		/* 检查登录路由 */
		authGroup.PUT("/checkin", controllers.UserCheckInHandler)

		/* 退出登录路由 */
		authGroup.DELETE("/signout", controllers.UserSignOutHandler)

		/* 验证路由 */
		authGroup.GET(
			"/validate",
			middlewares.JWTValidator,
			func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{"code": "10040", "message": "JWT 验证通过"})
			},
		)
	}
}
