package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"naotodoserver/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type CheckInHandlerV1DTO struct {
	JWT string
}

func CheckInHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto CheckInHandlerV1DTO

	// 获取参数
	if dto.JWT = ctx.Query("jwt"); dto.JWT == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10021,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 验证用户 JWT 是否过期
	iUserJWTClaims, err := utils.ParseUserJWT(dto.JWT)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10022,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}
	if utils.IsUserJWTExpiredByClaims(&iUserJWTClaims) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10023,
			Message: "用户凭证已过期,请重新登录",
			Data:    "",
		})
		return
	}

	// 若用户 JWT 未过期,检查数据库是否有对应的 session 记录
	var session models.Session
	result := core.DB.Where(&models.Session{JWT: dto.JWT}).First(&session)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10024,
			Message: "用户凭证已过期,请重新登录",
			Data:    nil,
		})
		return
	}

	// 检查用户是否到了需要再次登录的时间
	if session.ExpiresAt.Before(time.Now()) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10025,
			Message: "用户凭证已过期,请重新登录",
			Data:    nil,
		})
		return
	}

	// 重新签发用户 JWT,并更新数据库对应的 session 记录
	newJWT, err := utils.GenerateUserJWT(utils.UserJWTClaimsProfile{
		Id:        iUserJWTClaims.Profile.Id,
		Avatar:    iUserJWTClaims.Profile.Avatar,
		Email:     iUserJWTClaims.Profile.Email,
		Nickname:  iUserJWTClaims.Profile.Nickname,
		Role:      iUserJWTClaims.Profile.Role,
		CreatedAt: iUserJWTClaims.Profile.CreatedAt,
	})
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10026,
			Message: "签发用户 JWT 失败",
			Data:    err.Error(),
		})
		return
	}
	var sessionCond models.Session
	sessionCond.JWT = dto.JWT
	sessionCond.UpdatedAt = time.Time(time.Now())
	result = core.DB.Where(&models.Session{JWT: dto.JWT}).UpdateColumns(&sessionCond)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    10027,
			Message: "更新用户 JWT 失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回数据
	apis.Success(ctx, apis.ResponseData{
		Code:    10020,
		Message: "检入成功",
		Data:    newJWT,
	})
}
