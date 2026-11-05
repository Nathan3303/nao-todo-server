package middlewares

import (
	"fmt"

	authApp "naotodoserver/application/auth"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// logRateLimitBlocked 记录一次限流拒绝（C7：10051 响应计数依据）。
// 只落 key 的 SHA-256 摘要与桶名，绝不落明文 key / userId / token / 邮箱（C5/C6）。
// @param key 已计数的限流键
// @param tag 桶名（sync-ip / sync-pull / sync-push ...）
// @param limit 当前桶配额
func logRateLimitBlocked(key string, tag string, limit int64) {
	logging.WarnWithFields(map[string]any{
		"event":   "ratelimit.blocked",
		"key_tag": logging.KeyTag(key),
		"bucket":  tag,
		"limit":   limit,
	}, "rate limit blocked")
}

// UserRateLimiter 按登录用户维度限流（SEC-07），必须挂在 JWTValidator 之后。
// 与 RateLimiter 共用同一固定窗口计数实现，仅 Key 改为 user:<id>:<tag>：
// 不随 IP 变化（NAT / 多设备不会互相拖累），也无法靠换 IP 绕过。
// key 不含 token / 邮箱，仅 userId 与桶名（#88 / C6）。
// @param auth 认证应用（提供 RateLimit）
// @param limit 每分钟配额
// @param tag 桶名（同一用户不同用途互不影响）
func UserRateLimiter(auth authApp.AuthApp, limit int64, tag string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 取当前用户（鉴权中间件已注入）
		userId := iCtx.GetUserId(ctx.Request.Context())
		if userId <= 0 {
			controllers.Failure(ctx, types.ResponseData{
				Code:    10041,
				Message: "用户凭证验证失败",
			})
			ctx.Abort()
			return
		}
		// 2. 按用户维度检查
		key := fmt.Sprintf("user:%d:%s", userId, tag)
		if err := auth.RateLimit(ctx, key, limit); err != nil {
			logRateLimitBlocked(key, tag, limit)
			controllers.Failure(ctx, types.ResponseData{
				Code:    10051,
				Message: "请求失败",
				Error:   "请求次数超出单位时间内次数",
			})
			ctx.Abort()
			return
		}
		// 3. next
		ctx.Next()
	}
}

func RateLimiter(auth authApp.AuthApp, limit int64, tag string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 构建 Key
		key := ctx.ClientIP() + ":" + tag
		// 2. 检查
		err := auth.RateLimit(ctx, key, limit)
		if err != nil {
			logRateLimitBlocked(key, tag, limit)
			controllers.Failure(ctx, types.ResponseData{
				Code:    10051,
				Message: "请求失败",
				Error:   "请求次数超出单位时间内次数",
			})
			ctx.Abort()
			return
		}
		// 3. next
		ctx.Next()
	}
}
