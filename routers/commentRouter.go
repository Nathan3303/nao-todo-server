package routers

import (
	commentapis "naotodoserver/apis/comment"
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func CommentRouterInit(router *gin.RouterGroup) {
	commentRouter := router.Group("/comment", userapis.ValidateHandlerV1)
	{
		commentRouter.POST("/", commentapis.CreateCommentHandlerV1)
		commentRouter.PUT("/:commentId", commentapis.UpdateCommentHandlerV1)
		commentRouter.DELETE("/:commentId", commentapis.DeleteCommentHandlerV1)
	}
	commentsRouter := router.Group("/comments", userapis.ValidateHandlerV1)
	{
		commentsRouter.GET("/", commentapis.GetCommentsHandlerV1)
	}
}
