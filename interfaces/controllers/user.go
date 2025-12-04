package controllers

import (
	"naotodoserver/application/user"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

/*
 * Update user nickname handler
 * 更新用户昵称（10050）
 */
func UpdateUserNicknameHandler(ctx *gin.Context) {
	// 1. 获取参数
	var req types.UpdateUserNicknameReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10051,
			Message: "参数错误",
		})
		return
	}
	// 2. 验证参数
	if len(req.Nickname) <= 2 || len(req.Nickname) > 32 {
		Failure(ctx, types.ResponseData{
			Code:    10052,
			Message: "昵称长度必须在 2-32 个字符之间",
		})
		return
	}
	// 3. 调用用户服务 - 更新用户昵称
	res, err := user.App.UpdateNickname(ctx.Request.Context(), req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10053,
			Message: "更新用户昵称失败 - " + err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10050,
		Message: "更新用户昵称成功",
		Data:    res,
	})
}

/*
 * Get user profile handler
 * 获取用户详情（10060）
 */
func GetUserProfileHandler(ctx *gin.Context) {
	// 1. 调用用户服务 - 获取用户详情
	res, err := user.App.GetProfile(ctx.Request.Context())
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10065,
			Message: "获取用户详情失败 - " + err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10060,
		Message: "获取用户详情成功",
		Data:    res,
	})
}

/*
 * Update user password handler
 * 更新用户密码（10070）
 */
func UpdateUserPasswordHandler(ctx *gin.Context) {
	// 1. 获取参数
	var req types.UpdateUserPasswordReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10071,
			Message: "参数错误",
		})
		return
	}
	// 2. 验证参数
	if len(req.NewPassword) < 8 || len(req.NewPassword) > 32 {
		Failure(ctx, types.ResponseData{
			Code:    10073,
			Message: "新密码长度必须在 8-32 个字符之间",
		})
		return
	}
	// 3. 调用用户服务 - 更新用户密码
	res, err := user.App.UpdatePassword(ctx.Request.Context(), req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10074,
			Message: "更新用户密码失败 - " + err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10070,
		Message: "更新用户密码成功",
		Data:    res,
	})
}

/*
 * Update user avatar handler
 * 更新用户头像（10080）
 */
func UpdateUserAvatarHandler(ctx *gin.Context) {
	// panic("unimplemented")
	// 1. 获取参数
	var req types.UpdateUserAvatarReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10081,
			Message: "参数错误",
		})
		return
	}
	res := &types.UpdateUserAvatarRes{}
	// 2. 判断是否通过 AvatarURL 更新头像
	if req.AvatarURL != "" {
		// 是：则调用用户服务 - 更新用户头像 URL
		res, err = user.App.UpdateAvatar(ctx.Request.Context(), req)
	} else {
		// 否：则调用用户服务 - 更新用户头像文件
		res, err = user.App.UpdateAvatarByFile(ctx)
	}
	// 3. 判断处理结果
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10082,
			Message: "更新用户头像失败 - " + err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10080,
		Message: "更新用户头像成功",
		Data:    res,
	})
}
