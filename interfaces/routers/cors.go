package routers

import (
	"regexp"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// corsExactOrigins 精确白名单：现网正式域 + 本地开发端口。
// 一条不得删减，且**不受 APP_DEBUG 影响**（生产行为只增不减）。
var corsExactOrigins = []string{
	"http://localhost:5173",
	"http://localhost:5174",
	"https://todo.nathanao.space",
	"https://todobe.nathanao.space",
}

// corsPreviewOriginPattern 本项目 Vercel 预览域：
//
//	https://naotodo-<分支或 PR slug>-nathan3303s-projects.vercel.app
//
// 只认本 team 的 `nathan3303s-projects` 后缀，换 team / 项目名时改这一处。
// 不放行任意 `*.vercel.app`（任何 Vercel 用户都能注册该类域名，AllowCredentials
// 为 true 时等于对全世界开放）；正则锚定整串 `^…$`，`…vercel.app.evil.com` 不会命中。
var corsPreviewOriginPattern = regexp.MustCompile(
	`^https://naotodo-[a-z0-9-]+-nathan3303s-projects\.vercel\.app$`,
)

// ipv4Octet 匹配 0–255 且不接受前导零的十进制八位组。
const ipv4Octet = `(?:25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])`

// corsLANOriginPattern 局域网来源：`http://<私网 IPv4>:<5173|5174>`，
// 覆盖 192.168.0.0/16、10.0.0.0/8、172.16.0.0/12 三个私网网段。
// **仅 APP_DEBUG=true 时生效**（生产不放开局域网）；锚定整串，端口必须在白名单内。
var corsLANOriginPattern = regexp.MustCompile(
	`^http://(?:` +
		`192\.168\.` + ipv4Octet + `\.` + ipv4Octet + // 192.168.x.x
		`|10\.` + ipv4Octet + `\.` + ipv4Octet + `\.` + ipv4Octet + // 10.x.x.x
		`|172\.(?:1[6-9]|2[0-9]|3[01])\.` + ipv4Octet + `\.` + ipv4Octet + // 172.16–31.x.x
		`):(?:5173|5174)$`,
)

// isAllowedOrigin 是 CORS 放行判定的**唯一来源**（禁止在别处再写一份）：
//
//  1. 精确白名单（始终生效）
//  2. 本项目 Vercel 预览域（始终生效，与 debug 无关）
//  3. 局域网私网网段 + 白名单端口（仅 debug 生效）
//
// 只返回布尔值：AllowCredentials 为 true 时「回显精确 origin」由中间件完成，不能用 "*"。
func isAllowedOrigin(origin string, debug bool) bool {
	if origin == "" {
		return false
	}
	for _, exact := range corsExactOrigins {
		if origin == exact {
			return true
		}
	}
	if corsPreviewOriginPattern.MatchString(origin) {
		return true
	}
	return debug && corsLANOriginPattern.MatchString(origin)
}

// newCORSMiddleware 构造 CORS 中间件；AllowMethods / AllowHeaders / ExposeHeaders
// 与改动前保持一致（本单不放宽任何方法或请求头）。
func newCORSMiddleware(debug bool) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return isAllowedOrigin(origin, debug)
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Device-Id"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
