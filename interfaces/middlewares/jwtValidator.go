package middlewares

import (
	"errors"
	"net/http"
	"strings"

	authApp "naotodoserver/application/auth"
	domerr "naotodoserver/domain/errors"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/logging"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// 允许用 ?token= 传递凭证的路由（SEC-06 / #67）：
//   - 头像：浏览器 <img> / 移动端 <image> 无法自定义请求头
//   - SSE：EventSource 无法自定义请求头（当前无客户端消费者，保留既有语义）
//
// 其余路由一律只认 Authorization，避免 7 天会话 token 落进 URL / 反代日志 / 截图 / 分享。
const (
	avatarRoutePrefix = "/static/uploads/avatars/"
	sseRoutePrefix    = "/api/sse/"
)

// getJwtString 提取请求凭证（SEC-06 / #67）。
// 顺序：Authorization: Bearer 优先；仅头像 / SSE 路由才回退 ?token=。
// 旧实现有两处问题：① query 优先 ⇒ 攻击者塞一个坏 query 就能让合法 header 请求失败；
// ② 所有路由都接受 query token ⇒ 凭证可经 URL 泄露。
func getJwtString(ctx *gin.Context) string {
	if token := bearerTokenFromHeader(ctx.GetHeader("Authorization")); token != "" {
		return token
	}
	if isQueryTokenAllowed(ctx) {
		return strings.TrimSpace(ctx.Query("token"))
	}
	return ""
}

// bearerTokenFromHeader 解析 RFC 7235 Bearer 凭证。
// scheme 大小写不敏感（兼容 "Bearer xxx" / "bearer xxx"）；容忍多余空白。
func bearerTokenFromHeader(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// isQueryTokenAllowed 判断当前请求是否允许用 ?token= 传凭证。
// 用 Gin **路由解析后的模板**（ctx.FullPath）做前缀匹配，而不是原始 URL.Path：
//   - 模板来自路由注册表，不含用户可控的 `..`、大小写或百分号编码变体，
//     无法借构造路径把 query token 带到其它路由；
//   - 未匹配任何路由（404）时模板为空 ⇒ 一律不允许。
func isQueryTokenAllowed(ctx *gin.Context) bool {
	route := ctx.FullPath()
	if route == "" {
		return false
	}
	return strings.HasPrefix(route, avatarRoutePrefix) ||
		strings.HasPrefix(route, sseRoutePrefix)
}

// 鉴权失败码型（DEF-143 / ADR 2026-11-10）
//
// 凭证类（10041）只能由**显式证据**触发：JWT 解析失败 / 过期 / 会话不存在或无效。
// 其余一切错误（含 DB/连接类与未知错误）一律回 HTTP 503 + 50300：
//   - 不得再折叠成 10041 —— 客户端会据此清认证并踢用户回登录页；
//   - 未知错误也按「不登出」处理（fail-safe：宁可少登出，不误登出）。
const (
	authFailureCode         = 10041
	authFailureMessage      = "用户凭证验证失败"
	storeUnavailableCode    = 50300
	storeUnavailableMessage = "服务暂时不可用，请稍后重试"
)

// event 名（告警口径，稳定关键字，不得随文案变动）
const (
	eventAuthRejected      = "auth.credential_rejected"
	eventSessionLookupFail = "auth.session_lookup_failed"
)

func JWTValidator(auth authApp.AuthApp) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 验证 JWT
		jwtString := getJwtString(ctx)
		userId, err := auth.Validate(ctx, jwtString)
		if err != nil {
			// 不向客户端暴露内部错误细节（会话不存在、token 过期等）
			if errors.Is(err, domerr.ErrTokenExpired) || errors.Is(err, domerr.ErrSessionNotFound) {
				logging.WarnWithFields(map[string]any{
					"event":      eventAuthRejected,
					"request_id": iCtx.GetRequestId(ctx.Request.Context()),
					"path":       ctx.Request.URL.Path,
				}, "凭证校验未通过")
				controllers.Failure(ctx, types.ResponseData{
					Code:    authFailureCode,
					Message: authFailureMessage,
				})
				ctx.Abort()
				return
			}
			// 基础设施类失败：不登出，回 503（AC-2）+ error 级结构化事件（AC-5）
			reason := "unknown"
			if errors.Is(err, domerr.ErrStoreUnavailable) {
				reason = "infrastructure"
			}
			logging.ErrorWithFields(map[string]any{
				"event":      eventSessionLookupFail,
				"request_id": iCtx.GetRequestId(ctx.Request.Context()),
				"user_id":    int64(userId),
				"path":       ctx.Request.URL.Path,
				"reason":     reason,
				"error":      err.Error(),
			}, "鉴权依赖失败（不折叠为凭证失效）")
			// 先写 user_id：访问日志（status>=500 自动 error）需要按用户归属（AC-7）
			ctx.Request = ctx.Request.WithContext(
				iCtx.SetUserId(ctx.Request.Context(), int64(userId)),
			)
			controllers.FailureByHttpStatus(ctx, http.StatusServiceUnavailable, types.ResponseData{
				Code:    storeUnavailableCode,
				Message: storeUnavailableMessage,
			})
			ctx.Abort()
			return
		}
		if userId <= 0 {
			controllers.Failure(ctx, types.ResponseData{
				Code:    10042,
				Message: "用户凭证验证失败",
			})
			ctx.Abort()
			return
		}
		// 2. 写入用户信息和 token 到上下文
		ctx.Request = ctx.Request.WithContext(
			iCtx.SetToken(
				iCtx.SetUserId(ctx.Request.Context(), int64(userId)),
				jwtString,
			),
		)
		// 3. 检测通过，继续处理请求
		ctx.Next()
	}
}
