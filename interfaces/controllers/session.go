package controllers

import (
	"errors"
	"strconv"

	authApp "naotodoserver/application/auth"
	authDto "naotodoserver/application/auth/dto"
	domerr "naotodoserver/domain/errors"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	authApp authApp.AuthApp
}

func NewSessionController(app authApp.AuthApp) *SessionController {
	return &SessionController{authApp: app}
}

// toSessionRes 将应用层会话列表项转换为会话响应
// @param items 应用层会话列表项
// @return 会话响应切片
func toSessionRes(items []*authDto.SessionItem) []*types.SessionRes {
	res := make([]*types.SessionRes, 0, len(items))
	for _, item := range items {
		res = append(res, &types.SessionRes{
			Id:         item.Id,
			DeviceId:   item.DeviceId,
			DeviceType: item.DeviceType,
			IP4:        item.IP4,
			Region:     item.Region,
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
			Current:    item.Current,
		})
	}
	return res
}

// ListUserSessions 获取用户现存会话列表控制器
// @code 1013x
func (c *SessionController) ListUserSessions(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID 与 token
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10131,
			Message: "用户未登录",
		})
		return
	}
	token := iCtx.GetToken(ctx.Request.Context())
	// 2. 调用认证服务 - 获取会话列表
	sessions, err := c.authApp.ListSessions(ctx.Request.Context(), userId, token)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10132,
			Message: "获取会话列表失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10130,
		Message: "获取会话列表成功",
		Data:    toSessionRes(sessions),
	})
}

// LogoutUserSession 下线指定会话控制器
// @code 1014x
func (c *SessionController) LogoutUserSession(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID（路由已过 JWT 校验，userId 恒有效；此处兜底与另两个 handler 保持一致）
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10143,
			Message: "用户未登录",
		})
		return
	}
	// 2. 解析会话 ID
	sessionId, err := strconv.ParseInt(ctx.Param("sessionId"), 10, 64)
	if err != nil || sessionId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10141,
			Message: "会话 ID 无效",
		})
		return
	}
	// 3. 调用认证服务 - 下线指定会话
	err = c.authApp.LogoutSession(ctx.Request.Context(), userId, sessionId)
	if err != nil {
		if errors.Is(err, domerr.ErrSessionNotFound) {
			Failure(ctx, types.ResponseData{
				Code:    10142,
				Message: "会话不存在或已下线",
			})
			return
		}
		Failure(ctx, types.ResponseData{
			Code:    10143,
			Message: "下线失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10140,
		Message: "下线成功",
	})
}

// LogoutOtherSessions 退出其他全部设备控制器
// @code 1015x
func (c *SessionController) LogoutOtherSessions(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID 与 token
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10151,
			Message: "用户未登录",
		})
		return
	}
	token := iCtx.GetToken(ctx.Request.Context())
	// 2. 调用认证服务 - 退出其他全部设备
	err := c.authApp.LogoutOtherSessions(ctx.Request.Context(), userId, token)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10152,
			Message: "退出其他设备失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10150,
		Message: "已退出其他全部设备",
	})
}
