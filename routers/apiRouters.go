package routers

import (
	"03.project-template/controllers/api"
	"github.com/gin-gonic/gin"
)

func ApiRoutersInit(router *gin.Engine) {
	apiRouter := router.Group("/api")

	// user
	userApiRouter := apiRouter.Group("/user")
	{
		userApiRouter.GET("/test", func(ctx *gin.Context) { ctx.String(200, "test") })
		userApiRouter.POST("/nickname", api.UserNicknameHandler)
		userApiRouter.POST("/password", api.UserPasswordHandler)
		userApiRouter.POST("/avatar", api.UserAvatarHandler)
	}

	// todo
	todoApiRouter := apiRouter.Group("/todo")
	{
		todoApiRouter.POST("/", func(ctx *gin.Context) {})
		todoApiRouter.DELETE("/", func(ctx *gin.Context) {})
		todoApiRouter.PUT("/", func(ctx *gin.Context) {})
		todoApiRouter.GET("/", func(ctx *gin.Context) {})
		// todoApiRouter.GET("/todos", func(ctx *gin.Context) {})
		todoApiRouter.GET("/duplicate", func(ctx *gin.Context) {})
		// todoApiRouter.DELETE("/todos", func(ctx *gin.Context) {})
	}

	// project
	projectApiRouter := apiRouter.Group("/project")
	{
		projectApiRouter.POST("/", func(ctx *gin.Context) {})
		projectApiRouter.DELETE("/", func(ctx *gin.Context) {})
		projectApiRouter.PUT("/", func(ctx *gin.Context) {})
		projectApiRouter.GET("/", func(ctx *gin.Context) {})
		// projectApiRouter.GET("/projects", func(ctx *gin.Context) {})
	}

	// tag
	tagApiRouter := apiRouter.Group("/tag")
	{
		tagApiRouter.POST("/", func(ctx *gin.Context) {})
		tagApiRouter.DELETE("/", func(ctx *gin.Context) {})
		tagApiRouter.PUT("/", func(ctx *gin.Context) {})
		tagApiRouter.GET("/", func(ctx *gin.Context) {})
		// tagApiRouter.GET("/tags", func(ctx *gin.Context) {})
	}

	// event
	eventApiRouter := apiRouter.Group("/event")
	{
		eventApiRouter.POST("/", func(ctx *gin.Context) {})
		eventApiRouter.DELETE("/", func(ctx *gin.Context) {})
		eventApiRouter.PUT("/", func(ctx *gin.Context) {})
		eventApiRouter.GET("/", func(ctx *gin.Context) {})
		// eventApiRouter.GET("/events", func(ctx *gin.Context) {})
	}

	// comment
	commentApiRouter := apiRouter.Group("/comment")
	{
		commentApiRouter.POST("/", func(ctx *gin.Context) {})
		commentApiRouter.DELETE("/", func(ctx *gin.Context) {})
		commentApiRouter.PUT("/", func(ctx *gin.Context) {})
		// commentApiRouter.GET("/", func(ctx *gin.Context) {})
		// commentApiRouter.GET("/comments", func(ctx *gin.Context) {})
	}

	// auth
	authApiRouter := apiRouter.Group("/auth")
	{
		authApiRouter.POST("/signin", func(ctx *gin.Context) {})
		authApiRouter.POST("/signup", func(ctx *gin.Context) {})
		authApiRouter.GET("/checkin", func(ctx *gin.Context) {})
		authApiRouter.DELETE("/signout", func(ctx *gin.Context) {})
		// authApiRouter.Any("/", func(ctx *gin.Context) {})
	}
}
