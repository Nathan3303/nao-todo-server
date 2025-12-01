package controllers

import (
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

/*
 * UserSignInHandler
 * 用户登录控制器
 */
func UserSignInHandler(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	signInReq := types.UserSignInReq{}
	err := ctx.ShouldBind(&signInReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10011,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// @step 2. 调用用户服务 - 登录
	signInRes, err := auth..SignIn(ctx, signInReq)
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

func UserSignUpHandler(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	var signUpReq = types.UserSignUpReq{}
	err := ctx.ShouldBind(&signUpReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10001,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// @step 2. 调用用户服务 - 注册
	signUpRes, err := user.UserService.SignUp(ctx, signUpReq)
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
		Data:    signUpRes,
	})
}

func UserCheckInHandler(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	var checkInReq = types.UserCheckInReq{}
	err := ctx.ShouldBind(&checkInReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10021,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// @step 2. 调用用户服务 - 签到
	checkInRes, err := user.UserService.CheckIn(ctx, checkInReq)
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

func UserSignOutHandler(ctx *gin.Context) {
	// @step 1. 绑定请求参数
	var signOutReq = types.UserSignOutReq{}
	err := ctx.ShouldBind(&signOutReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10031,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// @step 2. 调用用户服务 - 登出
	signOutRes, err := user.UserService.SignOut(ctx, signOutReq)
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
		Data:    signOutRes,
	})
}

func UpdateProfileHandler(ctx *gin.Context) {
	// 1. 绑定请求参数
	var req = types.UpdateProfileReq{}
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10051,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 2. 调用用户服务 - 更新昵称
	updateProfileRes, err := user.UserService.UpdateProfile(ctx, req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    10052,
			Message: "更新个人信息失败",
			Data:    err.Error(),
		})
		return
	}

	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    10050,
		Message: "更新个人信息成功",
		Data:    updateProfileRes,
	})
}
