package controllers

import (
	"naotodoserver/application/project"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

/*
 * Get Project Handler
 * 获取清单（20000）
 */
func GetProjectHandler(ctx *gin.Context) {
	// 1. 绑定请求体
	var req types.GetProjectReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20001,
			Message: "参数错误",
			Data:    "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 获取清单
	res, err := project.App.Get(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20002,
			Message: "获取清单失败",
			Data:    err.Error(),
		})
		return
	}
	// 3. 实体转换响应体并返回
	Success(ctx, types.ResponseData{
		Code:    20000,
		Message: "获取清单成功",
		Data:    res,
	})
}

/*
 * Create Project Handler
 * 创建清单（20010）
 */
func CreateProjectHandler(ctx *gin.Context) {
	// 1. 获取参数
	var req types.CreateProjectReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20011,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// 2. 调用应用函数 - 创建清单
	res, err := project.App.Create(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20012,
			Message: "创建清单失败",
			Data:    err.Error(),
		})
		return
	}
	// 3. 实体转换响应体并返回
	Success(ctx, types.ResponseData{
		Code:    20010,
		Message: "创建清单成功",
		Data:    res,
	})
}

/*
 * Update Project Handler
 * 更新清单（20020）
 */
func UpdateProjectHandler(ctx *gin.Context) {
	// 1. 获取参数
	var req types.UpdateProjectReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20021,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}
	// 2. 获取清单 ID
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20022,
			Message: "参数错误",
			Data:    "清单 ID 不能为空",
		})
		return
	}
	// 3. 调用应用函数 - 更新清单
	res, err := project.App.Update(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20023,
			Message: "更新清单失败",
			Data:    err.Error(),
		})
		return
	}
	// 4. 实体转换响应体并返回
	Success(ctx, types.ResponseData{
		Code:    20020,
		Message: "更新清单成功",
		Data:    res,
	})
}

/*
 * Delete Project Handler
 * 删除清单（20030）
 */
func DeleteProjectHandler(ctx *gin.Context) {
	// 1. 获取清单 ID
	var req types.DeleteProjectReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20031,
			Message: "参数错误",
			Data:    "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 删除清单
	res, err := project.App.Delete(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20032,
			Message: "删除清单失败",
			Data:    err.Error(),
		})
		return
	}
	// 3. 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20030,
		Message: "删除清单成功",
		Data:    res,
	})
}

/*
 * Restore Project Handler
 * 恢复清单（20040）
 */
func RestoreProjectHandler(ctx *gin.Context) {
	// 1. 获取清单 ID
	var req types.RestoreProjectReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20041,
			Message: "参数错误",
			Data:    "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 恢复清单
	res, err := project.App.Restore(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20042,
			Message: "恢复清单失败",
			Data:    err.Error(),
		})
		return
	}
	// 3. 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20040,
		Message: "恢复清单成功",
		Data:    res,
	})
}

/*
 * Archive Project Handler
 * 归档清单（20050）
 */
func ArchiveProjectHandler(ctx *gin.Context) {
	// 1. 获取清单 ID
	var req types.ArchiveProjectReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20051,
			Message: "参数错误",
			Data:    "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 归档清单
	res, err := project.App.Archive(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20052,
			Message: "归档清单失败",
			Data:    err.Error(),
		})
		return
	}
	// 3. 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20050,
		Message: "归档清单成功",
		Data:    res,
	})
}

/*
 * Unarchive Project Handler
 * 取消归档清单（20060）
 */
func UnarchiveProjectHandler(ctx *gin.Context) {
	// 1. 获取清单 ID
	var req types.UnarchiveProjectReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20061,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 取消归档清单
	res, err := project.App.Unarchive(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20062,
			Message: "取消归档清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20060,
		Message: "取消归档清单成功",
		Data:    res,
	})
}

/*
 * List Project Handler
 * 获取清单列表（20070）
 */
func ListProjectHandler(ctx *gin.Context) {
	// 1. 调用应用函数 - 获取清单列表
	res, err := project.App.List(ctx.Request.Context())
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20071,
			Message: "获取清单列表失败",
			Error:   err.Error(),
		})
		return
	}
	// 2. 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20070,
		Message: "获取清单列表成功",
		Data:    res,
	})
}

/*
 * Get Project Preference Handler
 * 获取清单偏好（20080）
 */
func GetProjectPreferenceHandler(ctx *gin.Context) {
	// 1. 获取清单 ID
	var req types.GetProjectPreferenceReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20081,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 获取清单偏好
	res, err := project.App.GetPreference(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20082,
			Message: "获取清单偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 响应结果
	Success(ctx, types.ResponseData{
		Code:    20080,
		Message: "获取清单偏好成功",
		Data:    res,
	})
}

/*
 * Save Project Preference Handler
 * 保存清单偏好（20090）
 */
func SaveProjectPreferenceHandler(ctx *gin.Context) {
	// 1. 获取 ProjectId
	var req types.UpdateProjectPreferenceReq
	req.ProjectId = ctx.Param("projectId")
	if req.ProjectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20091,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 2. 获取更新参数
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20092,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 调用应用函数 - 保存清单偏好
	res, err := project.App.SavePreference(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20093,
			Message: "保存清单偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 响应结果
	Success(ctx, types.ResponseData{
		Code:    20090,
		Message: "保存清单偏好成功",
		Data:    res,
	})
}
