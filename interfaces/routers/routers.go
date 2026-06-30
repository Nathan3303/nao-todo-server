package routers

import (
	"naotodoserver/conf"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// securityHeaders 设置安全响应头（替代 nginx add_header）
func securityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("X-Frame-Options", "SAMEORIGIN")
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-XSS-Protection", "1; mode=block")
		ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		ctx.Next()
	}
}

// InitRouters 初始化路由
func InitRouters() *gin.Engine {
	// 是否启用 Debug模式
	if !conf.Conf.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 默认配置引擎
	router := gin.Default()

	// 禁用信任代理 IP
	err := router.SetTrustedProxies(nil)
	if err != nil {
		panic("设置信任代理 IP 失败: " + err.Error())
	}

	// 安全响应头（替代 nginx add_header）
	router.Use(securityHeaders())

	// 添加 Gzip 压缩（排除 SSE 路径，避免缓冲破坏实时推送）
	router.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithExcludedPaths([]string{"/api/sse/"}),
	))

	// 添加 CORS 中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://localhost:4173",
			"https://todo.nathanao.space",
			"https://todobe.nathanao.space",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// 	logging.Logger.
		// 		WithFields(map[string]any{"origin": origin}).
		// 		Info("CORS 允许来源")
		// 	return true
		// },
		MaxAge: 12 * time.Hour,
	}))

	// 配置静态文件服务器
	// 映射 /static/uploads 到本地 uploads 目录
	router.Static("/static/uploads", "./uploads")

	// 创建 API v1 路由组 - 应用 ClientInfo 和 RequestLogger 中间件
	v1 := router.Group(
		"/api",
		middlewares.ClientInfo,
		middlewares.RequestLogger(),
	)
	{
		v1.GET("/ping", controllers.PingHandler)
		UseAuthRouter(v1)
		UseUserRouter(v1)
		UseProjectRouter(v1)
		UseTagRouter(v1)
		UseTaskRouter(v1)
		UseEventRouter(v1)
		UseCommentRouter(v1)
		UsePomodoroRouter(v1)
		UseSSERouter(v1)
	}

	// 返回 Gin 引擎
	return router
}
