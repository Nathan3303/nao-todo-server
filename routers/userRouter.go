package routers

import (
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func UserRouterInit(router *gin.RouterGroup) {
	userRouter := router.Group("/user")
	{
		userRouter.POST("/signup", userapis.SignUpHandlerV1)
		userRouter.POST("/signin", userapis.SignInHandlerV1)
		userRouter.PUT("/checkin", userapis.CheckInHandlerV1)
		userRouter.DELETE("/signout", userapis.SignOutHandlerV1)
		userRouter.Any("/validate", userapis.ValidateHandlerV1)
	}
}
