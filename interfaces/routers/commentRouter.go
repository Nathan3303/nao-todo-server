package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UseCommentRouter 评论路由
func UseCommentRouter(router *gin.RouterGroup, ctrl *controllers.CommentController) {
	commentGroup := router.Group(
		"/comments",
		middlewares.RateLimiter(32, "comments"),
		middlewares.JWTValidator,
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
