package main

import (
	"fmt"
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/flags"
	"naotodoserver/routers"
	"naotodoserver/utils"
	"time"

	"github.com/gin-contrib/cors"
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

	// 自动迁移
	core.CheckAndExecuteAutoMigration()

	// 初始化 IP 解析工具
	utils.InitIpSearcher()

	// 初始化并运行 Gin
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/", func(ctx *gin.Context) {
		apis.Success(ctx, apis.ResponseData{Code: 200, Message: "Hello, Gin!", Data: nil})
	})
	routers.RoutersInit(router)                                                // 初始化路由
	router.Run(fmt.Sprintf("%s:%s", core.Config.Gin.Ip, core.Config.Gin.Port)) // 运行 Gin 服务

	// 测试 JWT 签发和解析
	// jwt, _ := utils.GenerateUserJWT(1, "nathan33")
	// fmt.Println(jwt)
	// userJWTClaims, err := utils.ParseUserJWT(jwt)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println(userJWTClaims)
	// }
}
