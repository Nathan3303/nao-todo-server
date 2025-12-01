package controllers

import (
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

func UpdateUserNicknameHandler(ctx *gin.Context) {
	// 1. 获取参数
	var req types.UpdateUserNicknameReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10001,
			Message: "参数错误",
		})
		return
	}
	// 2. 验证参数
	if len(req.Nickname) <= 2 || len(req.Nickname) > 32 {
		Failure(ctx, types.ResponseData{
			Code:    10001,
			Message: "昵称长度必须在 2-32 个字符之间",
		})
	}
	// 3. 调用用户服务

}
