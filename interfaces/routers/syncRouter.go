package routers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/conf"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// UseSyncRouter 配置数据同步路由（#88 / T656 C1-C3）。
//
// 中间件链：IP 粗桶 → JWT → 用户分桶。
//   - 前置 IP 桶（tag=sync-ip）是**未鉴权请求的地板**：限流挪到鉴权后就不再计数，
//     必须在验签/查库前拦住滥发，避免旧 token 持有者把 MySQL 打满；
//   - 后置 user 桶按 userId 计 pull / push 两桶（user:<id>:<tag>），
//     配额独立可配，pull 不会吃掉 push 配额。
//
// 响应码三态（C3）：有效未超限=原样；有效超限=10051；凭证无效=10041（无论是否超限）。
func UseSyncRouter(
	router *gin.RouterGroup,
	ctrl *controllers.SyncController,
	auth authApp.AuthApp,
) {
	limits := conf.Conf.Sync.RateLimit
	syncGroup := router.Group(
		"/sync",
		middlewares.RateLimiter(auth, limits.IPPerMin, "sync-ip"),
		middlewares.JWTValidator(auth),
	)
	{
		syncGroup.POST(
			"/push",
			middlewares.UserRateLimiter(auth, limits.PushPerMin, "sync-push"),
			ctrl.Push,
		)
		syncGroup.POST(
			"/pull",
			middlewares.UserRateLimiter(auth, limits.PullPerMin, "sync-pull"),
			ctrl.Pull,
		)
	}
}
