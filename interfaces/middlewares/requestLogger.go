package middlewares

import (
	"naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/logging"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 接口请求/响应日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 从上下文中获取客户端信息（ClientInfo 中间件已处理）
		clientInfo := context.GetClientInfo(c.Request.Context())
		// 检查客户端信息是否有效
		if clientInfo.IP4 == "" {
			clientInfo = context.ClientInfo{
				IP4:        c.ClientIP(),
				IPRegion:   "Unknown",
				DeviceType: "Unknown",
			}
		}

		// 继续处理请求
		c.Next()

		// 计算响应时间
		duration := time.Since(start)

		// 构建日志字段
		fields := map[string]interface{}{
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"status":          c.Writer.Status(),
			"duration":        duration.String(),
			"ip":              clientInfo.IP4,
			"ip_region":       clientInfo.IPRegion,
			"device_type":     clientInfo.DeviceType,
			"user_agent":      c.Request.UserAgent(),
			"content_type":    c.ContentType(),
			"content_length":  c.Request.ContentLength,
			"response_length": c.Writer.Size(),
		}

		// 获取用户ID（如果JWT验证通过）
		if userId := context.GetUserId(c.Request.Context()); userId > 0 {
			fields["user_id"] = userId
		}

		// 根据状态码决定日志级别
		switch {
		case c.Writer.Status() >= 500:
			logging.Logger.WithFields(fields).Error("API 服务器错误")
		case c.Writer.Status() >= 400:
			logging.Logger.WithFields(fields).Warn("API 客户端错误")
		default:
			logging.Logger.WithFields(fields).Info("API 请求完成")
		}
	}
}
