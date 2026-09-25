package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseAuthRouter(
	router *gin.RouterGroup,
	ctrl *controllers.AuthController,
	auth authApp.AuthApp,
) {
	authGroup := router.Group("/auth")
	{
		// 登录使用独立限流桶，避免被 checkin/signout 等操作挤占登录额度；其余接口共享一个桶
		authGroup.POST("/signin", middlewares.RateLimiter(auth, 8, "auth-signin"), ctrl.UserSignIn)
		authGroup.POST("/signup", middlewares.RateLimiter(auth, 8, "auth"), ctrl.UserSignUp)
		authGroup.PUT("/checkin", middlewares.RateLimiter(auth, 8, "auth"), ctrl.UserCheckIn)
		authGroup.DELETE("/signout", middlewares.RateLimiter(auth, 8, "auth"), ctrl.UserSignOut)
		authGroup.GET(
			"/validate",
			middlewares.RateLimiter(auth, 8, "auth"),
			middlewares.JWTValidator(auth),
			func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{"code": 10040, "message": "JWT 验证通过"})
			},
		)
	}
}
