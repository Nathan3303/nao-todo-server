package routers

import (
	"naotodoserver/application"
	"naotodoserver/conf"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// avatarRateLimit 头像读取配额（每个登录用户 / 60 秒，SEC-07）。
// 阈值取 120：首屏评论/成员列表可能批量加载不同头像（浏览器对头像有 immutable 缓存，
// 正常浏览远低于该值），同时抑制按 ID 批量枚举。
const avatarRateLimit int64 = 120

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

// bodyLimitOverrides 请求体上限的按路径覆盖表（SEC-01 / T561）。
// 路径必须与真实注册路径**逐字一致**：头像上传放宽到 MaxUploadBodyBytes（5 MB）。
// 注意：实际路由是 PUT /api/user/avatar（见 userRouter.go），不是 /api/users/avatar（F2）。
// @return 路径 → 上限（字节）
func bodyLimitOverrides() map[string]int64 {
	return map[string]int64{"/api/user/avatar": conf.Conf.Server.MaxUploadBodyBytes}
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

	// 请求体上限（SEC-01 / T561）：全局默认 2 MB，头像上传单独放宽 5 MB。
	// 必须在任何 ShouldBind / io.ReadAll 之前生效；超限 413，不静默截断。
	router.Use(middlewares.BodyLimit(
		conf.Conf.Server.MaxRequestBodyBytes,
		bodyLimitOverrides(),
	))

	// 添加 Gzip 压缩（排除 SSE 路径，避免缓冲破坏实时推送）
	router.Use(gzip.Gzip(
		gzip.DefaultCompression,
		gzip.WithExcludedPaths([]string{"/api/sse/"}),
	))

	// 请求关联标识（SEC-12）：只进日志，不进响应
	router.Use(middlewares.RequestId())

	// 添加 CORS 中间件（放行规则见 cors.go：精确白名单 + 预览域 + 局域网，仅 debug 放局域网）
	router.Use(newCORSMiddleware(conf.Conf.Server.Debug))

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
		sessionCtrl := controllers.NewSessionController(svc.Auth)
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
		UseUserRouter(v1, userCtrl, sessionCtrl, svc.Auth)
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
		middlewares.UserRateLimiter(svc.Auth, avatarRateLimit, "avatar"),
		userCtrl.GetAvatar,
	)

	// 返回 Gin 引擎
	return router
}
