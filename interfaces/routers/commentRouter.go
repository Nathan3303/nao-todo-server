package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// 注册评论路由组
func UseCommentRouter(router *gin.RouterGroup) {
	commentGroup := router.Group(
		"/comments",
		middlewares.RateLimiter(20, "comments"),
		middlewares.JWTValidator,
	)
	{
		// 获取评论列表
		commentGroup.GET("/", controllers.ListCommentHandler)

		// 获取评论详情
		commentGroup.GET("/:commentId", controllers.GetCommentHandler)

		// 创建评论
		commentGroup.POST("/", controllers.CreateCommentHandler)

		// 更新评论
		commentGroup.PUT("/:commentId", controllers.UpdateCommentHandler)

		// 删除评论
		commentGroup.DELETE(
			"/:commentId",
			controllers.DeleteCommentHandler,
		)
	}
}
