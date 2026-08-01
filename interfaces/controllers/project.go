package controllers

import (
	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	projectApp projectApp.ProjectApp
}

func NewProjectController(app projectApp.ProjectApp) *ProjectController {
	return &ProjectController{projectApp: app}
}

// toGetProjectRes 将应用层获取任务清单出参转换为获取任务清单响应
// @param output 应用层获取任务清单出参
// @return 获取任务清单响应
func toGetProjectRes(output *projectDto.GetProjectRes) *types.GetProjectRes {
	res := &types.GetProjectRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.Name = output.Name
	res.Description = output.Description
	res.SortId = output.SortId
	res.ArchivedAt = output.ArchivedAt
	res.DeactivedAt = output.DeactivedAt
	return res
}

// toGetProjectResList 将应用层任务清单列表出参转换为任务清单列表响应
// @param outputList 应用层任务清单列表出参
// @return 任务清单列表响应
func toGetProjectResList(outputList []*projectDto.GetProjectRes) []*types.GetProjectRes {
	resList := make([]*types.GetProjectRes, 0, len(outputList))
	for _, output := range outputList {
		resList = append(resList, toGetProjectRes(output))
	}
	return resList
}

// toCreateProjectInput 将创建任务清单请求转换为应用层入参
// @param req 创建任务清单请求
// @return 应用层创建任务清单入参
func toCreateProjectInput(req *types.CreateProjectReq) *projectDto.CreateProjectReq {
	return &projectDto.CreateProjectReq{
		Name:        req.Name,
		Description: req.Description,
	}
}

// toCreateProjectRes 将应用层创建任务清单出参转换为创建任务清单响应
// @param output 应用层创建任务清单出参
// @return 创建任务清单响应
func toCreateProjectRes(output *projectDto.CreateProjectRes) *types.CreateProjectRes {
	res := &types.CreateProjectRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.Name = output.Name
	res.Description = output.Description
	res.SortId = output.SortId
	res.ArchivedAt = output.ArchivedAt
	res.DeactivedAt = output.DeactivedAt
	return res
}

// toUpdateProjectInput 将更新任务清单请求转换为应用层入参
// @param req 更新任务清单请求
// @return 应用层更新任务清单入参
func toUpdateProjectInput(req *types.UpdateProjectReq) *projectDto.UpdateProjectReq {
	return &projectDto.UpdateProjectReq{
		Name:        req.Name,
		Description: req.Description,
		SortId:      req.SortId,
	}
}

// toBatchUpdateProjectInput 将批量更新任务清单请求转换为应用层入参
// @param req 批量更新任务清单请求
// @return 应用层批量更新任务清单入参
func toBatchUpdateProjectInput(req *types.BatchUpdateProjectReq) *projectDto.BatchUpdateProjectReq {
	projects := make([]*projectDto.BatchUpdateProjectItem, 0, len(req.Projects))
	for _, project := range req.Projects {
		projects = append(projects, &projectDto.BatchUpdateProjectItem{
			Id:          project.Id,
			Name:        project.Name,
			Description: project.Description,
			SortId:      project.SortId,
		})
	}
	return &projectDto.BatchUpdateProjectReq{Projects: projects}
}

// toBatchUpdateProjectRes 将应用层批量更新任务清单出参转换为批量更新任务清单响应
// @param output 应用层批量更新任务清单出参
// @return 批量更新任务清单响应
func toBatchUpdateProjectRes(
	output *projectDto.BatchUpdateProjectRes,
) *types.BatchUpdateProjectRes {
	return &types.BatchUpdateProjectRes{
		UpdatedCount: output.UpdatedCount,
		Projects:     toGetProjectResList(output.Projects),
	}
}

// toGetProjectPreferenceRes 将应用层获取任务清单偏好出参转换为获取任务清单偏好响应
// @param output 应用层获取任务清单偏好出参
// @return 获取任务清单偏好响应
func toGetProjectPreferenceRes(
	output *projectDto.GetProjectPreferenceRes,
) *types.GetProjectPreferenceRes {
	res := &types.GetProjectPreferenceRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.ProjectId = output.ProjectId
	res.ViewType = output.ViewType
	res.GetOptions = output.GetOptions
	res.Columns = output.Columns
	return res
}

// toUpdateProjectPreferenceInput 将更新任务清单偏好请求转换为应用层入参
// @param req 更新任务清单偏好请求
// @return 应用层更新任务清单偏好入参
func toUpdateProjectPreferenceInput(
	req *types.UpdateProjectPreferenceReq,
) *projectDto.UpdateProjectPreferenceReq {
	return &projectDto.UpdateProjectPreferenceReq{
		ViewType:   req.ViewType,
		GetOptions: req.GetOptions,
		Columns:    req.Columns,
	}
}

// GetProject 根据清单 ID 获取清单接入点
// @code 2000x
func (c *ProjectController) GetProject(ctx *gin.Context) {
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
	res, err := c.projectApp.Get(ctx.Request.Context(), projectId)
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
		Data:    toGetProjectRes(res),
	})
}

// CreateProject 创建清单接入点
// @code 2001x
func (c *ProjectController) CreateProject(ctx *gin.Context) {
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
	res, err := c.projectApp.Create(ctx.Request.Context(), toCreateProjectInput(&createProjectReq))
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
		Data:    toCreateProjectRes(res),
	})
}

// UpdateProject 更新清单接入点
// @code 2002x
func (c *ProjectController) UpdateProject(ctx *gin.Context) {
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
	err = c.projectApp.Update(
		ctx.Request.Context(),
		projectId,
		toUpdateProjectInput(&updateProjectReq),
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

// DeleteProject 删除清单接入点
// @code 2003x
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
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
	err := c.projectApp.Delete(ctx.Request.Context(), projectId)
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

// RestoreProject 恢复清单接入点
// @code 2004x
func (c *ProjectController) RestoreProject(ctx *gin.Context) {
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
	err := c.projectApp.Restore(ctx.Request.Context(), projectId)
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

// ArchiveProject 归档清单接入点
// @code 2005x
func (c *ProjectController) ArchiveProject(ctx *gin.Context) {
	// 获取清单 ID
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
	err := c.projectApp.Archive(ctx.Request.Context(), projectId)
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

// UnarchiveProject 取消归档清单接入点
// @code 2006x
func (c *ProjectController) UnarchiveProject(ctx *gin.Context) {
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
	err := c.projectApp.Unarchive(ctx.Request.Context(), projectId)
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

// ListProject 获取清单列表接入点
// @code 2007x
func (c *ProjectController) ListProject(ctx *gin.Context) {
	// 获取清单列表
	res, err := c.projectApp.List(ctx.Request.Context())
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
		Data:    toGetProjectResList(res),
	})
}

// GetProjectPreference 获取清单偏好接入点
// @code 2008x
func (c *ProjectController) GetProjectPreference(ctx *gin.Context) {
	// 获取清单 ID
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
	res, err := c.projectApp.GetPreference(ctx.Request.Context(), projectId)
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
		Data:    toGetProjectPreferenceRes(res),
	})
}

// BatchUpdateProjects 批量更新清单接入点
// @code 2010x
func (c *ProjectController) BatchUpdateProjects(ctx *gin.Context) {
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
	res, err := c.projectApp.BatchUpdate(ctx.Request.Context(), toBatchUpdateProjectInput(&req))
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
		Data:    toBatchUpdateProjectRes(res),
	})
}

// SaveProjectPreference 保存清单偏好接入点
// @code 2009x
func (c *ProjectController) SaveProjectPreference(ctx *gin.Context) {
	// 获取清单 ID
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
	err = c.projectApp.SavePreference(
		ctx.Request.Context(),
		projectId,
		toUpdateProjectPreferenceInput(&updatePreferenceReq),
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
