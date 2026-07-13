package controllers

import (
	userApp "naotodoserver/application/user"
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

// UpdateUserNickname 更新用户昵称控制器
// @code 1005x
func (c *UserController) UpdateUserNickname(ctx *gin.Context) {
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
	if utils.RuneLength(req.Nickname) < 2 || utils.RuneLength(req.Nickname) > 20 {
		Failure(ctx, types.ResponseData{
			Code:    10052,
			Message: "昵称长度必须在 2-20 个字符之间",
		})
		return
	}
	// 3. 调用用户服务 - 更新用户昵称
	err = c.userApp.UpdateNickname(ctx.Request.Context(), req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10053,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10050,
		Message: "更新用户昵称成功",
	})
}

// GetUserProfile 获取用户详情控制器
// @code 1006x
func (c *UserController) GetUserProfile(ctx *gin.Context) {
	// 1. 调用用户服务 - 获取用户详情
	res, err := c.userApp.GetProfile(ctx.Request.Context())
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

// UpdateUserPassword 更新用户密码控制器
// @code 1007x
func (c *UserController) UpdateUserPassword(ctx *gin.Context) {
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
	if utils.RuneLength(req.NewPassword) < 8 || utils.RuneLength(req.NewPassword) > 32 {
		Failure(ctx, types.ResponseData{
			Code:    10073,
			Message: "新密码长度必须在 8-32 个字符之间",
		})
		return
	}
	// 3. 调用用户服务 - 更新用户密码
	err = c.userApp.UpdatePassword(ctx.Request.Context(), req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10074,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10070,
		Message: "更新用户密码成功",
	})
}

// UpdateUserAvatar 更新用户头像控制器
// @code 1008x
func (c *UserController) UpdateUserAvatar(ctx *gin.Context) {
	// 1. 判断请求类型
	contentType := ctx.ContentType()

	var req types.UpdateUserAvatarReq
	var res *types.UpdateUserAvatarRes
	var err error

	// 2. 根据请求类型处理
	switch contentType {
	case "application/json", "text/plain;charset=utf-8":
		// JSON 请求：通过 URL 更新
		err = ctx.ShouldBind(&req)
		if err != nil {
			Failure(ctx, types.ResponseData{
				Code:    10081,
				Message: "参数错误",
			})
			return
		}

		if req.AvatarURL != "" {
			res, err = c.userApp.UpdateAvatar(ctx.Request.Context(), req)
		} else {
			Failure(ctx, types.ResponseData{
				Code:    10081,
				Message: "头像 URL 不能为空",
			})
			return
		}
	case "multipart/form-data":
		// 表单请求：通过文件上传更新
		res, err = c.userApp.UpdateAvatarByFile(ctx, ctx.Request.Context())
	default:
		Failure(ctx, types.ResponseData{
			Code:    10081,
			Message: "不支持的请求类型",
		})
		return
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

// DeactiveUser 禁用用户控制器
// @code 1009x
func (c *UserController) DeactiveUser(ctx *gin.Context) {
	// 1. 获取参数
	var req types.DeactiveUserReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10091,
			Message: "参数错误",
		})
		return
	}
	// 2. 调用用户服务 - 禁用用户
	err = c.userApp.DeactiveUser(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10092,
			Message: "禁用用户失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10090,
		Message: "禁用用户成功",
	})
}

// ActiveUser 启用用户控制器
// @code 1010x
func (c *UserController) ActiveUser(ctx *gin.Context) {
	// 1. 获取参数
	var req types.ActiveUserReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10101,
			Message: "参数错误",
		})
		return
	}
	// 2. 调用用户服务 - 启用用户
	err = c.userApp.ActiveUser(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10102,
			Message: "启用用户失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10100,
		Message: "启用用户成功",
	})
}

// GetUserConfig 获取用户配置控制器
// @code 1011x
func (c *UserController) GetUserConfig(ctx *gin.Context) {
	res, err := c.userApp.GetConfig(ctx.Request.Context())
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10111,
			Message: "获取用户配置失败 - " + err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    10110,
		Message: "获取用户配置成功",
		Data:    res,
	})
}

// UpdateUserConfig 更新用户配置控制器
// @code 1012x
func (c *UserController) UpdateUserConfig(ctx *gin.Context) {
	var req types.UpdateUserConfigReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10121,
			Message: "参数错误",
		})
		return
	}
	// if req.Appearance != "auto" && req.Appearance != "light" && req.Appearance != "dark" {
	// 	Failure(ctx, types.ResponseData{
	// 		Code:    10122,
	// 		Message: "外观设置只能为 auto、light 或 dark",
	// 	})
	// 	return
	// }
	err = c.userApp.UpdateConfig(ctx.Request.Context(), req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10123,
			Message: "更新用户配置失败 - " + err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    10120,
		Message: "更新用户配置成功",
	})
}
