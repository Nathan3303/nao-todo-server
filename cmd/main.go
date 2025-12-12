package main

import (
	"fmt"
	"naotodoserver/conf"
	"naotodoserver/infrastructure"
	"naotodoserver/interfaces/initialize"
)

func main() {
	// @step 1. 加载配置文件
	conf.InitConfig()

	// @step 2. 加载 Infrastructure 层
	infrastructure.Init()

	// @step 3. 加载路由
	router := initialize.InitRouters()
	err := router.Run(fmt.Sprintf("%s:%s", conf.Conf.Server.Ip, conf.Conf.Server.Port))
	if err != nil {
		panic("服务器启动失败 - " + err.Error())
	}
}
