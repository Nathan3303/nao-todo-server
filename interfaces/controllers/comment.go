package controllers

import (
	taskApp "naotodoserver/application/task"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type CommentController struct {
	taskApp taskApp.TaskCommentApp
}

func NewCommentController(app taskApp.TaskCommentApp) *CommentController {
	return &CommentController{taskApp: app}
}

// GetComment 获取评论详情控制器
// @code 6000x
func (c *CommentController) GetComment(ctx *gin.Context) {
	// 1. 获取评论 ID
	commentId := ctx.Param("commentId")
	if commentId == "" {
		Failure(ctx, types.ResponseData{
			Code:    60001,
			Message: "评论 ID 不能为空",
		})
		return
	}
	// 2. 调用服务层获取评论详情
	comment, err := c.taskApp.GetTaskCommentById(ctx.Request.Context(), commentId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60002,
			Message: "获取评论详情失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回评论详情
	Success(ctx, types.ResponseData{
		Code:    60000,
		Message: "获取评论详情成功",
		Data:    toTaskCommentRes(comment),
	})
}

// CreateComment 新增评论控制器
// @code 6001x
func (c *CommentController) CreateComment(ctx *gin.Context) {
	// 1. 绑定请求参数
	var req types.CreateTaskCommentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60011,
			Message: "绑定请求参数失败",
			Error:   err.Error(),
		})
		return
	}
	// 2. 调用服务层新增评论
	res, err := c.taskApp.CreateTaskComment(ctx.Request.Context(), toCreateTaskCommentReq(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60012,
			Message: "新增评论失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回评论 ID
	Success(ctx, types.ResponseData{
		Code:    60010,
		Message: "新增评论成功",
		Data:    toTaskCommentRes(res),
	})
}

// UpdateComment 更新评论控制器
// @code 6002x
func (c *CommentController) UpdateComment(ctx *gin.Context) {
	// 1. 获取评论 ID
	commentId := ctx.Param("commentId")
	if commentId == "" {
		Failure(ctx, types.ResponseData{
			Code:    60021,
			Message: "评论 ID 不能为空",
		})
		return
	}
	// 2. 绑定请求参数
	var req types.UpdateTaskCommentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60022,
			Message: "绑定请求参数失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 调用服务层更新评论
	err := c.taskApp.UpdateTaskComment(
		ctx.Request.Context(),
		commentId,
		toUpdateTaskCommentReq(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60023,
			Message: "更新评论失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回评论 ID
	Success(ctx, types.ResponseData{
		Code:    60020,
		Message: "更新评论成功",
		Data:    commentId,
	})
}

// DeleteComment 删除评论控制器
// @code 6003x
func (c *CommentController) DeleteComment(ctx *gin.Context) {
	// 1. 获取评论 ID
	commentId := ctx.Param("commentId")
	if commentId == "" {
		Failure(ctx, types.ResponseData{
			Code:    60031,
			Message: "评论 ID 不能为空",
		})
		return
	}
	// 2. 调用服务层删除评论
	err := c.taskApp.DeleteTaskComment(ctx.Request.Context(), commentId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60032,
			Message: "删除评论失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回评论 ID
	Success(ctx, types.ResponseData{
		Code:    60030,
		Message: "删除评论成功",
		Data:    commentId,
	})
}

// ListComment 获取评论列表控制器
// @code 6004x
func (c *CommentController) ListComment(ctx *gin.Context) {
	// 1. 获取待办任务 ID
	taskId := ctx.Query("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    60041,
			Message: "待办任务 ID 不能为空",
		})
		return
	}
	// 2. 调用服务层获取评论列表
	comments, err := c.taskApp.ListTaskComments(ctx.Request.Context(), taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    60042,
			Message: "获取评论列表失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回评论列表
	Success(ctx, types.ResponseData{
		Code:    60040,
		Message: "获取评论列表成功",
		Data:    toTaskCommentResList(comments),
	})
}
