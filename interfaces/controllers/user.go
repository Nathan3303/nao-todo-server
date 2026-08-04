package controllers

import (
	"mime"
	"net/http"
	"path/filepath"

	iCtx "naotodoserver/infrastructure/context"
	userApp "naotodoserver/application/user"
	userDto "naotodoserver/application/user/dto"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userApp userApp.UserApp
}

func NewUserController(app userApp.UserApp) *UserController {
	return &UserController{userApp: app}
}

// toUpdateNicknameInput 将更新用户昵称请求转换为应用层入参
// @param req 更新用户昵称请求
// @return 应用层更新用户昵称入参
func toUpdateNicknameInput(req types.UpdateUserNicknameReq) userDto.UpdateNicknameInput {
	return userDto.UpdateNicknameInput{
		Nickname: req.Nickname,
	}
}

// toUpdatePasswordInput 将更新用户密码请求转换为应用层入参
// @param req 更新用户密码请求
// @return 应用层更新用户密码入参
func toUpdatePasswordInput(req types.UpdateUserPasswordReq) userDto.UpdatePasswordInput {
	return userDto.UpdatePasswordInput{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
}

// toUpdateAvatarInput 将更新用户头像请求转换为应用层入参
// @param req 更新用户头像请求
// @return 应用层更新用户头像入参
func toUpdateAvatarInput(req types.UpdateUserAvatarReq) userDto.UpdateAvatarInput {
	return userDto.UpdateAvatarInput{
		AvatarURL: req.AvatarURL,
	}
}

// toGetProfileRes 将应用层获取用户个人信息出参转换为获取用户个人信息响应
// @param output 应用层获取用户个人信息出参
// @return 获取用户个人信息响应
func toGetProfileRes(output *userDto.GetProfileOutput) *types.GetUserProfileRes {
	return &types.GetUserProfileRes{
		Email:         output.Email,
		Nickname:      output.Nickname,
		Avatar:        output.Avatar,
		CreatedFrom:   output.CreatedFrom,
		Role:          output.Role,
		State:         output.State,
		DeactivedAt:   output.DeactivedAt,
		LastRestoreAt: output.LastRestoreAt,
		Config:        output.Config,
		CreatedAt:     output.CreatedAt,
		UpdatedAt:     output.UpdatedAt,
	}
}

// toUpdateAvatarRes 将应用层更新用户头像出参转换为更新用户头像响应
// @param output 应用层更新用户头像出参
// @return 更新用户头像响应
func toUpdateAvatarRes(output *userDto.UpdateAvatarOutput) *types.UpdateUserAvatarRes {
	return &types.UpdateUserAvatarRes{
		AvatarURL: output.AvatarURL,
	}
}

// toDeleteUserInput 将删除用户请求转换为应用层入参
// @param req 删除用户请求
// @return 应用层删除用户入参
func toDeleteUserInput(req types.DeleteUserReq) userDto.DeleteUserInput {
	return userDto.DeleteUserInput{
		Password: req.Password,
	}
}

// toRestoreUserInput 将激活用户请求转换为应用层入参
// @param req 激活用户请求
// @return 应用层激活用户入参
func toRestoreUserInput(req types.RestoreUserReq) userDto.RestoreUserInput {
	return userDto.RestoreUserInput{
		Password: req.Password,
	}
}

// toGetConfigRes 将应用层获取用户配置出参转换为获取用户配置响应
// @param output 应用层获取用户配置出参
// @return 获取用户配置响应
func toGetConfigRes(output *userDto.GetConfigOutput) *types.GetUserConfigRes {
	return &types.GetUserConfigRes{
		Appearance: output.Appearance,
	}
}

// toUpdateConfigInput 将更新用户配置请求转换为应用层入参
// @param req 更新用户配置请求
// @return 应用层更新用户配置入参
func toUpdateConfigInput(req types.UpdateUserConfigReq) userDto.UpdateConfigInput {
	return userDto.UpdateConfigInput{
		Appearance: req.Appearance,
	}
}

// UpdateUserNickname 更新用户昵称控制器
// @code 1005x
func (c *UserController) UpdateUserNickname(ctx *gin.Context) {
	// 1. 获取参数
	var req types.UpdateUserNicknameReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10051, Message: "参数错误"})
		return
	}
	// 2. 验证参数
	if utils.RuneLength(req.Nickname) < 2 || utils.RuneLength(req.Nickname) > 20 {
		Failure(ctx, types.ResponseData{
			Code:    10052,
			Message: "昵称长度必须在 2-20 个字符之间",
		})
		return
	}
	// 3. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10053,
			Message: "用户未登录",
		})
		return
	}
	// 4. 调用用户服务 - 更新用户昵称
	err = c.userApp.UpdateNickname(ctx.Request.Context(), userId, toUpdateNicknameInput(req))
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10054, Message: err.Error()})
		return
	}
	// 5. 返回结果
	Success(ctx, types.ResponseData{Code: 10050, Message: "更新用户昵称成功"})
}

// GetUserProfile 获取用户详情控制器
// @code 1006x
func (c *UserController) GetUserProfile(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10061,
			Message: "用户未登录",
		})
		return
	}
	// 2. 调用用户服务 - 获取用户详情
	res, err := c.userApp.GetProfile(ctx.Request.Context(), userId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10065,
			Message: "获取用户详情失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10060,
		Message: "获取用户详情成功",
		Data:    toGetProfileRes(res),
	})
}

// UpdateUserPassword 更新用户密码控制器
// @code 1007x
func (c *UserController) UpdateUserPassword(ctx *gin.Context) {
	// 1. 获取参数
	var req types.UpdateUserPasswordReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10071, Message: "参数错误"})
		return
	}
	// 2. 验证参数
	if utils.RuneLength(req.NewPassword) < 8 || utils.RuneLength(req.NewPassword) > 32 {
		Failure(ctx, types.ResponseData{
			Code:    10073,
			Message: "新密码长度必须在 8-32 个字符之间",
		})
		return
	}
	// 3. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10074,
			Message: "用户未登录",
		})
		return
	}
	// 4. 调用用户服务 - 更新用户密码
	err = c.userApp.UpdatePassword(ctx.Request.Context(), userId, toUpdatePasswordInput(req))
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10075, Message: err.Error()})
		return
	}
	// 5. 返回结果
	Success(ctx, types.ResponseData{Code: 10070, Message: "更新用户密码成功"})
}

// UpdateUserAvatar 更新用户头像控制器
// @code 1008x
func (c *UserController) UpdateUserAvatar(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10083,
			Message: "用户未登录",
		})
		return
	}
	// 2. 判断请求类型
	contentType := ctx.ContentType()
	var req types.UpdateUserAvatarReq
	var res *types.UpdateUserAvatarRes
	var err error
	// 3. 根据请求类型处理
	switch contentType {
	case "application/json", "text/plain;charset=utf-8":
		// JSON 请求：通过 URL 更新
		err = ctx.ShouldBind(&req)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 10081, Message: "参数错误"})
			return
		}
		if req.AvatarURL != "" {
			var appRes *userDto.UpdateAvatarOutput
			appRes, err = c.userApp.UpdateAvatar(
				ctx.Request.Context(),
				userId,
				toUpdateAvatarInput(req),
			)
			if err == nil {
				res = toUpdateAvatarRes(appRes)
			}
		} else {
			Failure(ctx, types.ResponseData{Code: 10081, Message: "头像 URL 不能为空"})
			return
		}
	case "multipart/form-data":
		// 表单请求：通过文件上传更新
		fileHeader, formErr := ctx.FormFile("avatar")
		if formErr != nil {
			Failure(ctx, types.ResponseData{
				Code:    10082,
				Message: "更新用户头像失败",
				Error:   "文件上传失败 - " + formErr.Error(),
			})
			return
		}
		f, openErr := fileHeader.Open()
		if openErr != nil {
			Failure(ctx, types.ResponseData{
				Code:    10082,
				Message: "更新用户头像失败",
				Error:   "文件上传失败 - " + openErr.Error(),
			})
			return
		}
		defer f.Close()
		var appRes *userDto.UpdateAvatarOutput
		appRes, err = c.userApp.UpdateAvatarByFile(
			ctx.Request.Context(),
			userId,
			f,
			fileHeader.Filename,
			fileHeader.Size,
		)
		if err == nil {
			res = toUpdateAvatarRes(appRes)
		}
	default:
		Failure(ctx, types.ResponseData{Code: 10081, Message: "不支持的请求类型"})
		return
	}
	// 4. 判断处理结果
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10082,
			Message: "更新用户头像失败",
			Error:   err.Error(),
		})
		return
	}
	// 5. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10080,
		Message: "更新用户头像成功",
		Data:    res,
	})
}

// GetAvatar 获取头像文件控制器
// @code 1008x
// 需要登录（JWT 鉴权），登录用户可查看任意头像
func (c *UserController) GetAvatar(ctx *gin.Context) {
	// 1. 获取文件名
	filename := ctx.Param("filename")
	// 2. 打开头像文件
	f, err := c.userApp.GetAvatar(ctx.Request.Context(), filename)
	if err != nil {
		FailureByHttpStatus(ctx, http.StatusNotFound, types.ResponseData{
			Code:    10084,
			Message: "头像文件不存在",
		})
		return
	}
	defer f.Close()
	// 3. 设置响应头并输出文件内容
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx.DataFromReader(http.StatusOK, -1, contentType, f, map[string]string{
		"Cache-Control": "private, max-age=31536000, immutable",
	})
}

// DeleteUser 删除用户控制器（注销账户）
// @code 1009x
func (c *UserController) DeleteUser(ctx *gin.Context) {
	var req types.DeleteUserReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10091, Message: "参数错误"})
		return
	}
	// 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10093,
			Message: "用户未登录",
		})
		return
	}
	// 注入当前请求的会话令牌，供注销成功后删除当次 Token
	input := toDeleteUserInput(req)
	input.Token = iCtx.GetToken(ctx.Request.Context())
	err = c.userApp.DeleteUser(ctx.Request.Context(), userId, input)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10092,
			Message: "注销账户失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    10090,
		Message: "注销账户成功，账户数据将在 7 天后自动删除",
	})
}

// RestoreUser 激活用户控制器
// @code 1010x
func (c *UserController) RestoreUser(ctx *gin.Context) {
	// 1. 获取参数
	var req types.RestoreUserReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10101, Message: "参数错误"})
		return
	}
	// 2. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10103,
			Message: "用户未登录",
		})
		return
	}
	// 3. 调用用户服务 - 激活用户
	err = c.userApp.RestoreUser(ctx.Request.Context(), userId, toRestoreUserInput(req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10102,
			Message: "激活用户失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{Code: 10100, Message: "激活用户成功"})
}

// GetUserConfig 获取用户配置控制器
// @code 1011x
func (c *UserController) GetUserConfig(ctx *gin.Context) {
	// 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10112,
			Message: "用户未登录",
		})
		return
	}
	res, err := c.userApp.GetConfig(ctx.Request.Context(), userId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10111,
			Message: "获取用户配置失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    10110,
		Message: "获取用户配置成功",
		Data:    toGetConfigRes(res),
	})
}

// UpdateUserConfig 更新用户配置控制器
// @code 1012x
func (c *UserController) UpdateUserConfig(ctx *gin.Context) {
	var req types.UpdateUserConfigReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{Code: 10121, Message: "参数错误"})
		return
	}
	// 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    10122,
			Message: "用户未登录",
		})
		return
	}
	err = c.userApp.UpdateConfig(ctx.Request.Context(), userId, toUpdateConfigInput(req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10123,
			Message: "更新用户配置失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{Code: 10120, Message: "更新用户配置成功"})
}
