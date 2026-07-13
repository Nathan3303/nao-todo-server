package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

func UseAuthRouter(router *gin.RouterGroup, ctrl *controllers.AuthController) {
	authGroup := router.Group(
		"/auth",
		middlewares.RateLimiter(8, "auth"),
	)
	{
		authGroup.POST("/signin", ctrl.UserSignIn)
		authGroup.POST("/signup", ctrl.UserSignUp)
		authGroup.PUT("/checkin", ctrl.UserCheckIn)
		authGroup.DELETE("/signout", ctrl.UserSignOut)
		authGroup.GET(
			"/validate",
			middlewares.JWTValidator,
			func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{"code": 10040, "message": "JWT 验证通过"})
			},
		)
	}
}
