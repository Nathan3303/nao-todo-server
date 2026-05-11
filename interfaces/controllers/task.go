package controllers

import (
	"naotodoserver/application/task"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// GetTaskHandler 获取待办任务详情控制器
// @code 4000x
func GetTaskHandler(ctx *gin.Context) {
	// 1. 获取待办任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40001,
			Message: "任务 ID 无效",
		})
		return
	}
	// 2. 调用应用层获取任务信息
	res, err := task.App.GetTaskById(ctx.Request.Context(), taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40002,
			Message: err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40000,
		Message: "获取待办任务详细成功",
		Data:    res,
	})
}

// CreateTaskHandler 创建待办任务控制器
// @code 4001x
func CreateTaskHandler(ctx *gin.Context) {
	// 1. 绑定请求参数
	var req types.CreateTaskReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40011,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 2. 调用应用层创建任务
	res, err := task.App.CreateTask(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40012,
			Message: "创建待办任务失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40010,
		Message: "创建待办任务成功",
		Data:    res,
	})
}

// UpdateTaskHandler 更新待办任务控制器
// @code 4002x
func UpdateTaskHandler(ctx *gin.Context) {
	// 1. 绑定请求参数
	var req types.UpdateTaskReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40021,
			Message: err.Error(),
		})
		return
	}
	// 2. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40022,
			Message: "任务 ID 无效",
		})
		return
	}
	// 3. 调用应用层更新任务
	err = task.App.UpdateTask(ctx.Request.Context(), taskId, &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40023,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40020,
		Message: "更新待办任务成功",
		Data:    taskId,
	})
}

// DeleteTaskHandler 删除待办任务控制器
// @code 4003x
func DeleteTaskHandler(ctx *gin.Context) {
	// 1. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40031,
			Message: "任务 ID 无效",
		})
		return
	}
	// 2. 调用应用层删除任务
	err := task.App.DeleteTask(ctx.Request.Context(), taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40032,
			Message: err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40030,
		Message: "删除待办任务成功",
		Data:    taskId,
	})
}

// RestoreTaskHandler 恢复待办任务控制器
// @code 4004x
func RestoreTaskHandler(ctx *gin.Context) {
	// 1. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40041,
			Message: "任务 ID 无效",
		})
		return
	}
	// 2. 调用应用层恢复任务
	err := task.App.RestoreTask(ctx.Request.Context(), taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40042,
			Message: err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40040,
		Message: "恢复待办任务成功",
		Data:    taskId,
	})
}

// CopyTaskHandler 复制待办任务控制器
// @code 4006x
func CopyTaskHandler(ctx *gin.Context) {
	// 1. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40061,
			Message: "任务 ID 无效",
		})
		return
	}
	// 2. 调用应用层复制任务
	res, err := task.App.CopyTask(ctx.Request.Context(), taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40062,
			Message: err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40060,
		Message: "复制待办任务成功",
		Data:    res,
	})
}

// ListTaskHandler 获取待办任务列表控制器
// @code 4005x
func ListTaskHandler(ctx *gin.Context) {
	// 1. 绑定请求参数
	var req types.ListTaskReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40051,
			Message: err.Error(),
		})
		return
	}
	// 2. 调用应用层获取任务列表
	tasks, paginationRes, err := task.App.ListTask(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40052,
			Message: err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:       40050,
		Message:    "获取待办任务列表成功",
		Data:       tasks,
		Pagination: paginationRes,
	})
}
