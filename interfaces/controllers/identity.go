package controllers

import (
	authApp "naotodoserver/application/auth"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authApp authApp.AuthApp
}

func NewAuthController(app authApp.AuthApp) *AuthController {
	return &AuthController{authApp: app}
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
	signInRes, err := c.authApp.SignIn(ctx.Request.Context(), &req)
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
		Data:    signInRes,
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
	err = c.authApp.SignUp(ctx.Request.Context(), &req)
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
	checkInRes, err := c.authApp.CheckIn(ctx.Request.Context(), &req)
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
		Data:    checkInRes,
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
	err = c.authApp.SignOut(ctx.Request.Context(), &req)
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
