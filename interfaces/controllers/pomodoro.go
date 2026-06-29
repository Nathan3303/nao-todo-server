package controllers

import (
	"naotodoserver/application"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// CreatePomodoroHandler 创建专注记录控制器
// @code 7001x
func CreatePomodoroHandler(ctx *gin.Context) {
	var req types.CreatePomodoroReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70011,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}

	res, err := application.App.Pomodoro.Create(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70012,
			Message: "创建专注记录失败",
			Error:   err.Error(),
		})
		return
	}

	Success(ctx, types.ResponseData{
		Code:    70010,
		Message: "创建专注记录成功",
		Data:    res,
	})
}

// GetPomodoroHandler 获取专注记录详情控制器
// @code 7002x
func GetPomodoroHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70021,
			Message: "专注记录 ID 无效",
		})
		return
	}

	res, err := application.App.Pomodoro.Get(
		ctx.Request.Context(),
		&types.GetPomodoroReq{Id: id},
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70022,
			Message: err.Error(),
		})
		return
	}

	Success(ctx, types.ResponseData{
		Code:    70020,
		Message: "获取专注记录成功",
		Data:    res,
	})
}

// ListPomodoroHandler 获取专注记录列表控制器
// @code 7003x
func ListPomodoroHandler(ctx *gin.Context) {
	var req types.ListPomodoroReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70031,
			Message: err.Error(),
		})
		return
	}

	res, total, err := application.App.Pomodoro.List(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70032,
			Message: err.Error(),
		})
		return
	}

	Success(ctx, types.ResponseData{
		Code:    70030,
		Message: "获取专注记录列表成功",
		Data:    res,
		Pagination: &types.Pagination{
			Page:  req.Page,
			Limit: req.Limit,
			Total: total,
		},
	})
}
