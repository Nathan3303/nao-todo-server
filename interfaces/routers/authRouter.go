package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseAuthRouter(router *gin.RouterGroup) {
	authGroup := router.Group(
		"/auth",
		middlewares.RateLimiter(8, "auth"),
	)
	{
		authGroup.POST("/signin", controllers.UserSignInHandler)
		authGroup.POST("/signup", controllers.UserSignUpHandler)
		authGroup.PUT("/checkin", controllers.UserCheckInHandler)
		authGroup.DELETE("/signout", controllers.UserSignOutHandler)
		authGroup.GET(
			"/validate",
			middlewares.JWTValidator,
			func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{"code": 10040, "message": "JWT 验证通过"})
			},
		)
	}
}
