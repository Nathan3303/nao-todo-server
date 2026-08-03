package controllers

import (
	taskApp "naotodoserver/application/task"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	taskApp taskApp.TaskCheckItemApp
}

func NewEventController(app taskApp.TaskCheckItemApp) *EventController {
	return &EventController{taskApp: app}
}

// GetEvent 获取检查事项详情控制器
// @code 5000x
func (c *EventController) GetEvent(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50003,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取检查事项 ID
	eventId := ctx.Param("eventId")
	if eventId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50001,
			Message: "检查事项 ID 不能为空",
		})
		return
	}
	// 3. 调用服务层获取检查事项
	res, err := c.taskApp.GetTaskCheckItemById(
		ctx.Request.Context(),
		userId,
		eventId,
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50002,
			Message: "获取检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回检查事项
	Success(ctx, types.ResponseData{
		Code:    50000,
		Message: "获取检查事项成功",
		Data:    toGetTaskCheckItemRes(res),
	})
}

// CreateEvent 新增检查事项控制器
// @code 5001x
func (c *EventController) CreateEvent(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50013,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取请求参数
	var req types.CreateTaskCheckItemReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50011,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 调用服务层创建检查事项
	res, err := c.taskApp.CreateTaskCheckItem(
		ctx.Request.Context(),
		userId,
		toCreateTaskCheckItemReq(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50012,
			Message: "创建检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回检查事项
	Success(ctx, types.ResponseData{
		Code:    50010,
		Message: "创建检查事项成功",
		Data:    toCreateTaskCheckItemRes(res),
	})
}

// UpdateEvent 更新检查事项控制器
// @code 5002x
func (c *EventController) UpdateEvent(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50024,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取检查事项 ID
	eventId := ctx.Param("eventId")
	if eventId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50021,
			Message: "检查事项 ID 不能为空",
		})
		return
	}
	// 3. 获取请求参数
	var req types.UpdateTaskCheckItemReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50022,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 4. 调用服务层更新检查事项
	err = c.taskApp.UpdateTaskCheckItem(
		ctx.Request.Context(),
		userId,
		eventId,
		toUpdateTaskCheckItemReq(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50023,
			Message: "更新检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 5. 返回检查事项
	Success(ctx, types.ResponseData{
		Code:    50020,
		Message: "更新检查事项成功",
		Data:    eventId,
	})
}

// DeleteEvent 删除检查事项控制器
// @code 5003x
func (c *EventController) DeleteEvent(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50033,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取检查事项 ID
	eventId := ctx.Param("eventId")
	if eventId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50031,
			Message: "检查事项 ID 不能为空",
		})
		return
	}
	// 3. 调用服务层删除检查事项
	err := c.taskApp.DeleteTaskCheckItem(ctx.Request.Context(), userId, eventId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50032,
			Message: "删除检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回成功
	Success(ctx, types.ResponseData{
		Code:    50030,
		Message: "删除检查事项成功",
		Data:    eventId,
	})
}

// ListEvent 获取检查事项列表控制器
// @code 5004x
func (c *EventController) ListEvent(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50043,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取待办事项 ID
	taskId := ctx.Query("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    50041,
			Message: "待办事项 ID 不能为空",
		})
		return
	}
	// 3. 调用服务层获取检查事项列表
	res, err := c.taskApp.ListTaskCheckItems(ctx.Request.Context(), userId, taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50042,
			Message: "获取检查事项列表失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回检查事项列表
	Success(ctx, types.ResponseData{
		Code:    50040,
		Message: "获取检查事项列表成功",
		Data:    toGetTaskCheckItemResList(res),
	})
}

// BatchUpdateEvent 批量更新检查事项控制器
// @code 5006x
func (c *EventController) BatchUpdateEvent(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50063,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取请求参数
	var req types.BatchUpdateTaskCheckItemReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50061,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 调用服务层批量更新检查事项
	res, err := c.taskApp.BatchUpdateTaskCheckItems(
		ctx.Request.Context(),
		userId,
		toBatchUpdateTaskCheckItemReq(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    50062,
			Message: "批量更新检查事项失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回检查事项列表
	Success(ctx, types.ResponseData{
		Code:    50060,
		Message: "批量更新检查事项成功",
		Data:    toBatchUpdateTaskCheckItemRes(res),
	})
}
