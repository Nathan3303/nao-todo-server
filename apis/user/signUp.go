package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"

	"github.com/gin-gonic/gin"
)

type signUpHandlerV1DTO struct {
	Email    string `json:"email" binding:"email,required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

func SignUpHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto signUpHandlerV1DTO

	// 解析参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10001,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 检查数据库是否已有该用户
	var user models.User
	var result = core.DB.Where("email = ? and password = ?", dto.Email, dto.Password).First(&user)
	if result.Error != nil && user.ID != 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10002,
			Message: "用户已存在",
			Data:    nil,
		})
		return
	}

	// 创建用户
	user = models.User{
		Account:    dto.Email,
		Email:      dto.Email,
		Nickname:   dto.Nickname,
		Password:   dto.Password,
		Role:       "User",
		CreateForm: "WebClient",
		Avatar:     "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif?imageView2/1/w/80/h/80",
	}
	result = core.DB.Create(&user)
	if user.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10003,
			Message: "创建用户失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    10000,
		Message: "注册成功",
		Data:    nil,
	})
}
