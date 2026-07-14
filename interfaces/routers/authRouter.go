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
	authGroup := router.Group(
		"/auth",
		middlewares.RateLimiter(auth, 8, "auth"),
	)
	{
		authGroup.POST("/signin", ctrl.UserSignIn)
		authGroup.POST("/signup", ctrl.UserSignUp)
		authGroup.PUT("/checkin", ctrl.UserCheckIn)
		authGroup.DELETE("/signout", ctrl.UserSignOut)
		authGroup.GET(
			"/validate",
			middlewares.JWTValidator(auth),
			func(ctx *gin.Context) {
				ctx.JSON(200, gin.H{"code": 10040, "message": "JWT 验证通过"})
			},
		)
	}
}
