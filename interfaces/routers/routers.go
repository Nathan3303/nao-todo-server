package routers

import (
	"naotodoserver/application"
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
func InitRouters(svc *application.Services) *gin.Engine {
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
			"http://localhost:5273",
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

	// 创建 API v1 路由组 - 应用 ClientInfo 和 RequestLogger 中间件
	userCtrl := controllers.NewUserController(svc.User)
	v1 := router.Group(
		"/api",
		middlewares.ClientInfo,
		middlewares.RequestLogger(),
	)
	{
		v1.GET("/ping", controllers.PingHandler)
		authCtrl := controllers.NewAuthController(svc.Auth)
		projectCtrl := controllers.NewProjectController(svc.Project)
		taskCtrl := controllers.NewTaskController(svc.Task)
		eventCtrl := controllers.NewEventController(svc.TaskCheckItem)
		commentCtrl := controllers.NewCommentController(svc.TaskComment)
		tagCtrl := controllers.NewTagController(svc.Tag)
		pomodoroCtrl := controllers.NewPomodoroController(svc.Pomodoro)
		sseCtrl := controllers.NewSSEController()
		systemCtrl := controllers.NewSystemController()
		syncCtrl := controllers.NewSyncController(
			svc.Task,
			svc.TaskCheckItem,
			svc.TaskComment,
			svc.Project,
			svc.Tag,
			svc.Pomodoro,
		)

		UseAuthRouter(v1, authCtrl, svc.Auth)
		UseUserRouter(v1, userCtrl, svc.Auth)
		UseProjectRouter(v1, projectCtrl, svc.Auth)
		UseTagRouter(v1, tagCtrl, svc.Auth)
		UseTaskRouter(v1, taskCtrl, svc.Auth)
		UseEventRouter(v1, eventCtrl, svc.Auth)
		UseCommentRouter(v1, commentCtrl, svc.Auth)
		UsePomodoroRouter(v1, pomodoroCtrl, svc.Auth)
		UseSSERouter(v1, sseCtrl, svc.Auth)
		UseSystemRouter(v1, systemCtrl, svc.Auth)
		UseSyncRouter(v1, syncCtrl, svc.Auth)
	}

	// 配置头像文件访问路由（替代原先的静态目录直出）
	// 头像文件需登录（JWT 鉴权）后方可访问
	router.GET(
		"/static/uploads/avatars/:filename",
		middlewares.JWTValidator(svc.Auth),
		userCtrl.GetAvatar,
	)

	// 返回 Gin 引擎
	return router
}
