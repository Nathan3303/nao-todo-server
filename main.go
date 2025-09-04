package main

import (
	"fmt"

	"03.project-template/core"
	"03.project-template/flags"
	"03.project-template/ip"
	"github.com/gin-gonic/gin"
)

func main() {
	// 获取命令行参数
	flags.ParseFlags()

	// 读取配置文件
	core.ReadConfigurations()

	// 初始化日志工具
	var logConfig = core.Config.Log
	core.InitLogrus(logConfig.Dir, logConfig.App)

	// 链接数据库
	var dbConfig = core.Config.DB
	core.ConnectDB(dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Using)

	// 初始化 IP 解析工具
	ip.InitIpSearcher()

	// 初始化并运行 Gin
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(200, "%s", "Hello, gin")
	})
	var ginRunStr = fmt.Sprintf("%s:%s", core.Config.Gin.Ip, core.Config.Gin.Port)
	router.Run(ginRunStr)
}
