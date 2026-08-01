package controllers

import (
	authApp "naotodoserver/application/auth"
	authDto "naotodoserver/application/auth/dto"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authApp authApp.AuthApp
}

func NewAuthController(app authApp.AuthApp) *AuthController {
	return &AuthController{authApp: app}
}

// toSignInInput 将登录请求转换为应用层登录入参
// @param req 登录请求
// @return 应用层登录入参
func toSignInInput(req *types.SignInReq) *authDto.SignInInput {
	return &authDto.SignInInput{
		Email:    req.Email,
		Password: req.Password,
	}
}

// toSignInRes 将应用层登录出参转换为登录响应
// @param output 应用层登录出参
// @return 登录响应
func toSignInRes(output *authDto.SignInOutput) *types.SignInRes {
	return &types.SignInRes{
		Token:           output.Token,
		PendingDeletion: output.PendingDeletion,
		DeletedAt:       output.DeletedAt,
	}
}

// toSignUpInput 将注册请求转换为应用层注册入参
// @param req 注册请求
// @return 应用层注册入参
func toSignUpInput(req *types.SignUpReq) *authDto.SignUpInput {
	return &authDto.SignUpInput{
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	}
}

// toCheckInInput 将检入请求转换为应用层检入入参
// @param req 检入请求
// @return 应用层检入入参
func toCheckInInput(req *types.CheckInReq) *authDto.CheckInInput {
	return &authDto.CheckInInput{
		Token:      req.Token,
		DeviceType: req.DeviceType,
	}
}

// toCheckInRes 将应用层检入出参转换为检入响应
// @param output 应用层检入出参
// @return 检入响应
func toCheckInRes(output *authDto.CheckInOutput) *types.CheckInRes {
	return &types.CheckInRes{
		Token:           output.Token,
		PendingDeletion: output.PendingDeletion,
		DeletedAt:       output.DeletedAt,
	}
}

// toSignOutInput 将登出请求转换为应用层登出入参
// @param req 登出请求
// @return 应用层登出入参
func toSignOutInput(req *types.SignOutReq) *authDto.SignOutInput {
	return &authDto.SignOutInput{
		Token:      req.Token,
		DeviceType: req.DeviceType,
	}
}

// UserSignIn 用户登录控制器
// @code 1001x
func (c *AuthController) UserSignIn(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	req := types.SignInReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10011,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// @step 2. 调用用户服务 - 登录
	signInOutput, err := c.authApp.SignIn(ctx.Request.Context(), toSignInInput(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10012,
			Message: "登录失败",
			Data:    err.Error(),
		})
		return
	}
	// @step 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10010,
		Message: "登录成功",
		Data:    toSignInRes(signInOutput),
	})
}

// UserSignUp 用户注册控制器
// @code 1000x
func (c *AuthController) UserSignUp(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	var req = types.SignUpReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10001,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// @step 2. 调用用户服务 - 注册
	err = c.authApp.SignUp(ctx.Request.Context(), toSignUpInput(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10002,
			Message: "注册失败",
			Data:    err.Error(),
		})
		return
	}
	// @step 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10000,
		Message: "注册成功",
	})
}

// UserCheckIn 用户检入控制器
// @code 1002x
func (c *AuthController) UserCheckIn(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	var req = types.CheckInReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10021,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// @step 2. 调用用户服务 - 检入
	checkInOutput, err := c.authApp.CheckIn(ctx.Request.Context(), toCheckInInput(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10022,
			Message: "检入失败",
			Data:    err.Error(),
		})
		return
	}
	// @step 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10020,
		Message: "检入成功",
		Data:    toCheckInRes(checkInOutput),
	})
}

// UserSignOut 用户登出控制器
// @code 1003x
func (c *AuthController) UserSignOut(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	var req = types.SignOutReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10031,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// @step 2. 调用用户服务 - 登出
	err = c.authApp.SignOut(ctx.Request.Context(), toSignOutInput(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10032,
			Message: "登出失败",
			Data:    err.Error(),
		})
		return
	}
	// @step 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10030,
		Message: "登出成功",
	})
}
