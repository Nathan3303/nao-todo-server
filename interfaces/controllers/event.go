package controllers

import (
	"naotodoserver/application/event"
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
	res, err := event.App.GetEventById(ctx.Request.Context(), eventId)
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
	var req types.CreateEventReq
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
	res, err := event.App.CreateEvent(ctx.Request.Context(), &req)
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
	var req types.UpdateEventReq
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
	err = event.App.UpdateEvent(ctx.Request.Context(), eventId, &req)
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
	err := event.App.DeleteEvent(ctx.Request.Context(), eventId)
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
	res, err := event.App.ListEvent(ctx.Request.Context(), taskId)
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

// ResortEventHandler 重新排序两个检查事项控制器
// @code 5005x
func ResortEventHandler(ctx *gin.Context) {
	// 1. 获取请求参数
	var req types.ResortEventsReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50051,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 2. 调用服务层获取检查事项列表
	res, err := event.App.ResortEvents(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50052,
			Message: "重新排序检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 2. 返回检查事项列表
	Success(ctx, types.ResponseData{
		Code:    50050,
		Message: "重新排序检查事项成功",
		Data:    res,
	})
}
