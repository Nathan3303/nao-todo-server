package controllers

import (
	"naotodoserver/application/project"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// 根据清单 ID 获取清单接入点
// @code 2000x
func GetProjectHandler(ctx *gin.Context) {
	// 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20001,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 获取清单
	res, err := project.App.Get(ctx.Request.Context(), projectId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20002,
			Message: "获取清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 实体转换响应体并返回
	Success(ctx, types.ResponseData{
		Code:    20000,
		Message: "获取清单成功",
		Data:    res,
	})
}

// 创建清单接入点
// @code 2001x
func CreateProjectHandler(ctx *gin.Context) {
	// 获取参数
	var createProjectReq types.CreateProjectReq
	err := ctx.ShouldBind(&createProjectReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20011,
			Message: "创建清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 创建清单
	res, err := project.App.Create(ctx.Request.Context(), &createProjectReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20012,
			Message: "创建清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 实体转换响应体并返回
	Success(ctx, types.ResponseData{
		Code:    20010,
		Message: "创建清单成功",
		Data:    res,
	})
}

// 更新清单接入点
// @code 2002x
func UpdateProjectHandler(ctx *gin.Context) {
	// 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20022,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 获取参数
	var updateProjectReq types.UpdateProjectReq
	err := ctx.ShouldBind(&updateProjectReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20021,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 更新清单
	err = project.App.Update(
		ctx.Request.Context(),
		projectId,
		&updateProjectReq,
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20023,
			Message: "更新清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 实体转换响应体并返回
	Success(ctx, types.ResponseData{
		Code:    20020,
		Message: "更新清单成功",
		Data:    projectId,
	})
}

// 删除清单接入点
// @code 2003x
func DeleteProjectHandler(ctx *gin.Context) {
	// 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20031,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 删除清单
	err := project.App.Delete(ctx.Request.Context(), projectId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20032,
			Message: "删除清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20030,
		Message: "删除清单成功",
		Data:    projectId,
	})
}

// 恢复清单接入点
// @code 2004x
func RestoreProjectHandler(ctx *gin.Context) {
	// 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20041,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 恢复清单
	err := project.App.Restore(ctx.Request.Context(), projectId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20042,
			Message: "恢复清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20040,
		Message: "恢复清单成功",
		Data:    projectId,
	})
}

// 归档清单接入点
// @code 2005x
func ArchiveProjectHandler(ctx *gin.Context) {
	// 清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20051,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 归档清单
	err := project.App.Archive(ctx.Request.Context(), projectId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20052,
			Message: "归档清单失败",
			Error:   err.Error(),
		})
		return
	}
	// 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20050,
		Message: "归档清单成功",
		Data:    projectId,
	})
}

// 取消归档清单接入点
// @code 2006x
func UnarchiveProjectHandler(ctx *gin.Context) {
	// 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20061,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 2. 调用应用函数 - 取消归档清单
	err := project.App.Unarchive(ctx.Request.Context(), projectId)
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
		Data:    projectId,
	})
}

// 获取清单列表接入点
// @code 2007x
func ListProjectHandler(ctx *gin.Context) {
	// 获取清单列表
	res, err := project.App.List(ctx.Request.Context())
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20071,
			Message: "获取清单列表失败",
			Error:   err.Error(),
		})
		return
	}
	// 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20070,
		Message: "获取清单列表成功",
		Data:    res,
	})
}

// 获取清单偏好接入点
// @code 2008x
func GetProjectPreferenceHandler(ctx *gin.Context) {
	//. 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20081,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 获取清单偏好
	res, err := project.App.GetPreference(ctx.Request.Context(), projectId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20082,
			Message: "获取清单偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20080,
		Message: "获取清单偏好成功",
		Data:    res,
	})
}

// 批量更新清单接入点
// @code 2010x
func BatchUpdateProjectsHandler(ctx *gin.Context) {
	var req types.BatchUpdateProjectReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20101,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	res, err := project.App.BatchUpdate(ctx.Request.Context(), &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20102,
			Message: "批量更新清单失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    20100,
		Message: "批量更新清单成功",
		Data:    res,
	})
}

// 保存清单偏好接入点
// @code 2009x
func SaveProjectPreferenceHandler(ctx *gin.Context) {
	//. 获取清单 ID
	projectId := ctx.Param("projectId")
	if projectId == "" {
		Failure(ctx, types.ResponseData{
			Code:    20091,
			Message: "参数错误",
			Error:   "清单 ID 不能为空",
		})
		return
	}
	// 获取更新参数
	var updatePreferenceReq types.UpdateProjectPreferenceReq
	err := ctx.ShouldBindJSON(&updatePreferenceReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20092,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 保存清单偏好
	err = project.App.SavePreference(
		ctx.Request.Context(),
		projectId,
		&updatePreferenceReq,
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    20093,
			Message: "保存清单偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 返回成功响应
	Success(ctx, types.ResponseData{
		Code:    20090,
		Message: "保存清单偏好成功",
		Data:    projectId,
	})
}
