package main

import (
	"fmt"
	"naotodoserver/conf"
	"naotodoserver/infrastructure/container"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/interfaces/initialize"
)

func main() {
	// @step 1. 加载配置文件
	conf.InitConfig()

	// @step 2. 加载 Infrastructure 层
	dbs.InitMySQL()
	models.InitSnowflake(1)
	container.LoadDomain()

	// @step 3. 加载路由
	router := initialize.InitRouters()
	err := router.Run(fmt.Sprintf("%s:%s", conf.Conf.Server.Ip, conf.Conf.Server.Port))
	if err != nil {
		panic("服务器启动失败 - " + err.Error())
	}
}
