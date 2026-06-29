package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UseCommentRouter 评论路由
func UseCommentRouter(router *gin.RouterGroup) {
	commentGroup := router.Group(
		"/comments",
		middlewares.RateLimiter(32, "comments"),
		middlewares.JWTValidator,
	)
	{
		commentGroup.GET("/", controllers.ListCommentHandler)
		commentGroup.GET("/:commentId", controllers.GetCommentHandler)
		commentGroup.POST("/", controllers.CreateCommentHandler)
		commentGroup.PUT("/:commentId", controllers.UpdateCommentHandler)
		commentGroup.DELETE(
			"/:commentId",
			controllers.DeleteCommentHandler,
		)
	}
}
