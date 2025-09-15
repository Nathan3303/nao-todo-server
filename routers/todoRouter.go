package routers

import (
	todoapis "naotodoserver/apis/todo"
	userapis "naotodoserver/apis/user"

	"github.com/gin-gonic/gin"
)

func TodoRouterInit(router *gin.RouterGroup) {
	todoRouter := router.Group("/todo", userapis.ValidateHandlerV1)
	{
		todoRouter.GET("/:todoId", todoapis.GetTodoHandlerV1)
		todoRouter.POST("/", todoapis.CreateTodoHandlerV1)
		todoRouter.PUT("/:todoId", todoapis.UpdateTodoHandlerV1)
		todoRouter.DELETE("/:todoId", todoapis.DeleteTodoHandlerV1)
		todoRouter.PUT("/restore/:todoId", todoapis.RestoreTodoHandlerV1)
	}
	todosRouter := router.Group("/todos", userapis.ValidateHandlerV1)
	{
		todosRouter.GET("/", todoapis.GetTodosHandlerV1)
	}
}
