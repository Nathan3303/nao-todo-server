package routers

import (
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouters() *gin.Engine {
	// 创建 Gin 默认配置引擎
	router := gin.Default()

	// 添加 CORS 中间件
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

	// 创建 API v1 路由组 - 应用 ClientInfo 和 RequestLogger 中间件
	v1 := router.Group("/api", middlewares.ClientInfo, middlewares.RequestLogger())
	{
		v1.GET("/ping", controllers.PingHandler)
		UseAuthRouter(v1)
		UseUserRouter(v1)
		UseProjectRouter(v1)
		UseTagRouter(v1)
		UseTaskRouter(v1)
		UseEventRouter(v1)
		UseCommentRouter(v1)
	}

	// 设置可信代理 IP（负载均衡器或 CDN 的 IP 段）
	err := router.SetTrustedProxies([]string{
		"192.168.1.0/24",
		"10.0.0.0/8",
		"127.0.0.1",
	})
	if err != nil {
		panic(err)
	}

	// 返回 Gin 引擎
	return router
}
