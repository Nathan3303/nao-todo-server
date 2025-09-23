package main

import (
	"naotodoserver/core"
	"naotodoserver/flags"
	"naotodoserver/routers"
	"naotodoserver/utils"
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

	// 自动迁移
	core.CheckAndExecuteAutoMigration()

	// 初始化 IP 解析工具
	utils.InitIpSearcher()

	// 初始化并运行 Gin
	routers.RoutersInit() // 初始化路由
}
