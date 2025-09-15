package routers

import "github.com/gin-gonic/gin"

func RoutersInit(router *gin.Engine) {
	apiRouter := router.Group("/api")
	{
		UserRouterInit(apiRouter)
		ProjectRouterInit(apiRouter)
		TagRouterInit(apiRouter)
		TodoRouterInit(apiRouter)
	}
}
