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
			authGroup.GET("/validate", middlewares.JWTValidator)
		}

		// @step 3.2 用户路由组
		// userGroup := v1.Group("/user", middlewares.JWTValidator)
		// {
		// 	userGroup.PUT("/")
		// }

		// @step 3.2 项目路由组
		// projectGroup := v1.Group("/project", middlewares.JWTValidator)
		// {
		// 	projectGroup.GET("/:projectId", controllers.GetProjectHandler)
		// 	projectGroup.POST("/", controllers.CreateProjectHandler)
		// 	projectGroup.PUT("/:projectId", controllers.UpdateProjectHandler)
		// 	projectGroup.DELETE("/:projectId", controllers.DeleteProjectHandler)
		// 	projectGroup.POST("/restore/:projectId", controllers.RestoreProjectHandler)
		// }
		// projectsGroup := v1.Group("/projects", middlewares.JWTValidator)
		// {
		// 	projectsGroup.GET("/", controllers.GetProjectsHandler)
		// }
	}

	// @step 4. 返回 Gin 引擎
	return router
}
