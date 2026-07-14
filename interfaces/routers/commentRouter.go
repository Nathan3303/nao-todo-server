package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UseCommentRouter 评论路由
func UseCommentRouter(router *gin.RouterGroup, ctrl *controllers.CommentController, auth authApp.AuthApp) {
	commentGroup := router.Group(
		"/comments",
		middlewares.RateLimiter(auth, 32, "comments"),
		middlewares.JWTValidator(auth),
	)
	{
		commentGroup.GET("/", ctrl.ListComment)
		commentGroup.GET("/:commentId", ctrl.GetComment)
		commentGroup.POST("/", ctrl.CreateComment)
		commentGroup.PUT("/:commentId", ctrl.UpdateComment)
		commentGroup.DELETE(
			"/:commentId",
			ctrl.DeleteComment,
		)
	}
}
