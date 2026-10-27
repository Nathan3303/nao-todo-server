package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	iCtx "naotodoserver/infrastructure/context"

	"github.com/gin-gonic/gin"
)

// requestIdHeader 客户端可选传入的关联标识头（仅透传，不回写响应）
const requestIdHeader = "X-Request-Id"

// requestIdPattern 透传标识的字符白名单：字母/数字/点/下划线/连字符（防日志注入）
var requestIdPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestId 为每个请求准备一个关联标识（SEC-12）：
//   - 客户端传入合法 X-Request-Id ⇒ 透传（限长 + 字符白名单，非法则丢弃）
//   - 否则服务端随机生成 16 字节 hex
//
// 标识只写入上下文/日志字段，**不写入响应头或响应体**。
func RequestId() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestId := ctx.GetHeader(requestIdHeader)
		if !requestIdPattern.MatchString(requestId) {
			requestId = newRequestId()
		}
		ctx.Request = ctx.Request.WithContext(iCtx.SetRequestId(ctx.Request.Context(), requestId))
		ctx.Next()
	}
}

// newRequestId 生成服务端关联标识（16 字节随机 hex）
// @return 32 位十六进制字符串
func newRequestId() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
