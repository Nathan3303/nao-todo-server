package routers

import (
	"naotodoserver/apis"
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
		userRouter.PUT("/profile", userapis.ValidateHandlerV1, userapis.UpdateProfileHandlerV1)
		userRouter.PUT("/password", userapis.ValidateHandlerV1, userapis.UpdatePasswordHandlerV1)
		// ping
		userRouter.GET("/validate", userapis.ValidateHandlerV1, apis.DefaultHandler)
	}
}
