package api

import "github.com/gin-gonic/gin"

func UserNicknameHandler(ctx *gin.Context) {
	ctx.String(200, "[UserNicknameHandler]")
}

func UserPasswordHandler(ctx *gin.Context) {
	ctx.String(200, "[UserPasswordHandler]")
}

func UserAvatarHandler(ctx *gin.Context) {
	ctx.String(200, "[UserAvatarHandler]")
}
