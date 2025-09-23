package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"naotodoserver/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type SignInHandlerV1DTO struct {
	Email    string `json:"email" binding:"email,required"`
	Password string `json:"password" binding:"required"`
}

func SignInHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto SignInHandlerV1DTO

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10011,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 验证参数
	var userRaw = &models.User{}
	var result = core.DB.Where(&models.User{Email: dto.Email, Password: dto.Password}).Find(&userRaw)
	if userRaw.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10012,
			Message: "用户名或密码错误",
			Data:    nil,
		})
		return
	}

	// 签发用户 JWT
	var user = ToUserResponse(userRaw)
	jwt, err := utils.GenerateUserJWT(utils.UserJWTClaimsProfile{
		Id:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10013,
			Message: "签发用户 JWT 失败",
			Data:    err.Error(),
		})
		return
	}

	// 查找并更新现有的 Session 记录
	var session models.Session
	var jwtExpiresAt = time.Now().Add(time.Hour * 24)
	core.DB.Where(&models.Session{UserId: userRaw.ID}).Find(&session)
	if session.UserId == 0 {
		// 如果不存在则直接创建
		core.DB.Create(&models.Session{
			UserId:    userRaw.ID,
			JWT:       jwt,
			ExpiresAt: jwtExpiresAt,
		})
		core.DB.Where(&models.Session{UserId: userRaw.ID}).First(&session)
		if session.UserId == 0 {
			apis.Failure(ctx, apis.ResponseData{
				Code:    10014,
				Message: "创建用户 Session 失败",
				Data:    nil,
			})
			return
		}
	} else {
		// 存在则更新记录
		core.DB.Where(&session).UpdateColumns(&models.Session{
			JWT:       jwt,
			ExpiresAt: jwtExpiresAt,
		})
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    10010,
		Message: "登录成功",
		Data:    jwt,
	})
}
