package controllers

import (
	"naotodoserver/application"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// GetEventHandler 获取检查事项详情控制器
// @code 5000x
func GetEventHandler(ctx *gin.Context) {
	// 1. 获取检查事项 ID
	eventId := ctx.Param("eventId")
	if eventId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50001,
			Message: "检查事项 ID 不能为空",
		})
		return
	}
	// 2. 调用服务层获取检查事项
	res, err := application.App.Task.GetCheckItemById(ctx.Request.Context(), eventId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50002,
			Message: "获取检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回检查事项
	Success(ctx, types.ResponseData{
		Code:    50000,
		Message: "获取检查事项成功",
		Data:    res,
	})
}

// CreateEventHandler 新增检查事项控制器
// @code 5001x
func CreateEventHandler(ctx *gin.Context) {
	// 1. 获取请求参数
	var req types.CreateCheckItemReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50011,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 2. 调用服务层创建检查事项
	res, err := application.App.Task.CreateCheckItem(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50012,
			Message: "创建检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回检查事项
	Success(ctx, types.ResponseData{
		Code:    50010,
		Message: "创建检查事项成功",
		Data:    res,
	})
}

// UpdateEventHandler 更新检查事项控制器
// @code 5002x
func UpdateEventHandler(ctx *gin.Context) {
	// 1. 获取检查事项 ID
	eventId := ctx.Param("eventId")
	if eventId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50021,
			Message: "检查事项 ID 不能为空",
		})
		return
	}
	// 2. 获取请求参数
	var req types.UpdateCheckItemReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50022,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 调用服务层更新检查事项
	err = application.App.Task.UpdateCheckItem(ctx.Request.Context(), eventId, &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50023,
			Message: "更新检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回检查事项
	Success(ctx, types.ResponseData{
		Code:    50020,
		Message: "更新检查事项成功",
		Data:    eventId,
	})
}

// DeleteEventHandler 删除检查事项控制器
// @code 5003x
func DeleteEventHandler(ctx *gin.Context) {
	// 1. 获取检查事项 ID
	eventId := ctx.Param("eventId")
	if eventId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50031,
			Message: "检查事项 ID 不能为空",
		})
		return
	}
	// 2. 调用服务层删除检查事项
	err := application.App.Task.DeleteCheckItem(ctx.Request.Context(), eventId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50032,
			Message: "删除检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回成功
	Success(ctx, types.ResponseData{
		Code:    50030,
		Message: "删除检查事项成功",
		Data:    eventId,
	})
}

// ListEventHandler 获取检查事项列表控制器
// @code 5004x
func ListEventHandler(ctx *gin.Context) {
	// 1. 获取待办事项 ID
	taskId := ctx.Query("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50041,
			Message: "待办事项 ID 不能为空",
		})
		return
	}
	// 2. 调用服务层获取检查事项列表
	res, err := application.App.Task.ListCheckItems(ctx.Request.Context(), taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50042,
			Message: "获取检查事项列表失败",
			Error:   err.Error(),
		})
		return
	}
	// 2. 返回检查事项列表
	Success(ctx, types.ResponseData{
		Code:    50040,
		Message: "获取检查事项列表成功",
		Data:    res,
	})
}


// BatchUpdateEventHandler 批量更新检查事项控制器
// @code 5006x
func BatchUpdateEventHandler(ctx *gin.Context) {
	// 1. 获取请求参数
	var req types.BatchUpdateCheckItemReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50061,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 2. 调用服务层批量更新检查事项
	res, err := application.App.Task.BatchUpdateCheckItems(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50062,
			Message: "批量更新检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回检查事项列表
	Success(ctx, types.ResponseData{
		Code:    50060,
		Message: "批量更新检查事项成功",
		Data:    res,
	})
}
