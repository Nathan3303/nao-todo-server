package middlewares

import (
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/ip2region"
	"strings"

	"github.com/gin-gonic/gin"
)

func getClientType(userAgent string) string {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "micromessenger"):
		return "WeChat" // 微信内置浏览器
	case strings.Contains(ua, "alipayclient"):
		return "Alipay" // 支付宝
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "iphone") ||
		strings.Contains(ua, "ipad"):
		return "iOS"
	case strings.Contains(ua, "chrome"):
		return "Chrome"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		return "Safari"
	case strings.Contains(ua, "firefox"):
		return "Firefox"
	case strings.Contains(ua, "bot") ||
		strings.Contains(ua, "spider") ||
		strings.Contains(ua, "crawl"):
		return "Bot/Crawler"
	default:
		return "Unknown"
	}
}

func ClientInfo(ctx *gin.Context) {
	// 1. 获取客户端 IP 地址
	clientIp := ctx.ClientIP()
	// 2. 获取客户端 IP 地址区域
	clientRegion, err := ip2region.GetIp2RegionImpl().ParseIp(clientIp)
	if err != nil {
		clientRegion = "UnknownRegion"
	}
	// 3. 获取客户端设备类型
	deviceType := getClientType(ctx.Request.UserAgent())
	// 4. 构建 ClientInfo
	clientInfo := iCtx.ClientInfo{
		IP4:        clientIp,
		IPRegion:   clientRegion,
		DeviceType: deviceType,
	}
	// 2. 写入到上下文
	ctx.Request = ctx.Request.WithContext(
		iCtx.SetClientInfo(ctx.Request.Context(), clientInfo),
	)
	// 3. 继续处理请求
	ctx.Next()
}
