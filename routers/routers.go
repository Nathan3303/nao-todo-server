package routers

import (
	"fmt"
	"naotodoserver/core"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RoutersInit() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:4173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	apiRouter := router.Group("/api")
	{
		UserRouterInit(apiRouter)
		ProjectRouterInit(apiRouter)
		TagRouterInit(apiRouter)
		TodoRouterInit(apiRouter)
		EventRouterInit(apiRouter)
		CommentRouterInit(apiRouter)
	}

	router.Run(fmt.Sprintf("%s:%s", core.Config.Gin.Ip, core.Config.Gin.Port)) // 运行 Gin 服务
}
