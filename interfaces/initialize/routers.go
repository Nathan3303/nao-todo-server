package initialize

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouters() *gin.Engine {
	// @step 1. 创建 Gin 默认配置引擎
	router := gin.Default()

	// @step 2. 添加 CORS 中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:4173",
			"http://localhost",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// @step 3. 创建 API v1 路由组
	v1 := router.Group("/api")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.String(200, "Ping OK!")
		})

		// @step 3.1 验证路由组
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/signin", controllers.UserSignInHandler)
			authGroup.POST("/signup", controllers.UserSignUpHandler)
			authGroup.PUT("/checkin", controllers.UserCheckInHandler)
			authGroup.DELETE("/signout", controllers.UserSignOutHandler)
			authGroup.GET(
				"/validate",
				middlewares.JWTValidator,
				func(ctx *gin.Context) {
					ctx.JSON(200, gin.H{"code": "10040", "message": "JWT 验证通过"})
				},
			)
		}

		// @step 3.2 用户路由组
		userGroup := v1.Group("/user", middlewares.JWTValidator)
		{
			userGroup.GET("/profile", controllers.GetUserProfileHandler)
			userGroup.PUT("/nickname", controllers.UpdateUserNicknameHandler)
			userGroup.PUT("/password", controllers.UpdateUserPasswordHandler)
			userGroup.PUT("/avatar", controllers.UpdateUserAvatarHandler)
		}

		// @step 3.2 项目路由组
		projectGroup := v1.Group("/projects", middlewares.JWTValidator)
		{
			projectGroup.GET("/", controllers.ListProjectHandler)
			projectGroup.GET("/:projectId", controllers.GetProjectHandler)
			projectGroup.POST("/", controllers.CreateProjectHandler)
			projectGroup.PUT("/:projectId", controllers.UpdateProjectHandler)
			projectGroup.DELETE("/:projectId", controllers.DeleteProjectHandler)
			projectGroup.PUT("/restore/:projectId", controllers.RestoreProjectHandler)
			projectGroup.PUT("/archive/:projectId", controllers.ArchiveProjectHandler)
			projectGroup.PUT("/unarchive/:projectId", controllers.UnarchiveProjectHandler)
		}

		// @step 3.3 标签路由组
		tagGroup := v1.Group("/tags", middlewares.JWTValidator)
		{
			tagGroup.GET("/", controllers.ListTagHandler)
			tagGroup.GET("/:tagId", controllers.GetTagHandler)
			tagGroup.POST("/", controllers.CreateTagHandler)
			tagGroup.PUT("/:tagId", controllers.UpdateTagHandler)
			tagGroup.DELETE("/:tagId", controllers.DeleteTagHandler)
		}

		// @step 3.4 任务路由组
		taskGroup := v1.Group("/tasks", middlewares.JWTValidator)
		{
			taskGroup.GET("/", controllers.ListTaskHandler)
			taskGroup.GET("/:taskId", controllers.GetTaskHandler)
			taskGroup.POST("/", controllers.CreateTaskHandler)
			taskGroup.PUT("/:taskId", controllers.UpdateTaskHandler)
			taskGroup.DELETE("/:taskId", controllers.DeleteTaskHandler)
			taskGroup.PUT("/restore/:taskId", controllers.RestoreTaskHandler)
		}

		// @step 3.5 检查事项路由组
		eventGroup := v1.Group("/events", middlewares.JWTValidator)
		{
			eventGroup.GET("/", controllers.ListEventHandler)
			eventGroup.GET("/:eventId", controllers.GetEventHandler)
			eventGroup.POST("/", controllers.CreateEventHandler)
			eventGroup.PUT("/:eventId", controllers.UpdateEventHandler)
			eventGroup.DELETE("/:eventId", controllers.DeleteEventHandler)
		}

		// @step 3.6 评论路由组
		commentGroup := v1.Group("/comments", middlewares.JWTValidator)
		{
			commentGroup.GET("/", controllers.ListCommentHandler)
			commentGroup.GET("/:commentId", controllers.GetCommentHandler)
			commentGroup.POST("/", controllers.CreateCommentHandler)
			commentGroup.PUT("/:commentId", controllers.UpdateCommentHandler)
			commentGroup.DELETE("/:commentId", controllers.DeleteCommentHandler)
		}
	}

	// @step 4. 返回 Gin 引擎
	return router
}
