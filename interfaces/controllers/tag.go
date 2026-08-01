package controllers

import (
	tagApp "naotodoserver/application/tag"
	tagDto "naotodoserver/application/tag/dto"
	"naotodoserver/interfaces/types"
	"strings"

	"github.com/gin-gonic/gin"
)

type TagController struct {
	tagApp tagApp.TagApp
}

func NewTagController(app tagApp.TagApp) *TagController {
	return &TagController{tagApp: app}
}

// toGetTagRes 将应用层获取标签出参转换为获取标签响应
// @param output 应用层获取标签出参
// @return 获取标签响应
func toGetTagRes(output *tagDto.GetTagRes) *types.GetTagRes {
	res := &types.GetTagRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.Name = output.Name
	res.Description = output.Description
	res.Color = output.Color
	res.SortId = output.SortId
	return res
}

// toGetTagResList 将应用层标签列表出参转换为标签列表响应
// @param outputList 应用层标签列表出参
// @return 标签列表响应
func toGetTagResList(outputList []*tagDto.GetTagRes) []*types.GetTagRes {
	resList := make([]*types.GetTagRes, 0, len(outputList))
	for _, output := range outputList {
		resList = append(resList, toGetTagRes(output))
	}
	return resList
}

// toCreateTagInput 将创建标签请求转换为应用层入参
// @param req 创建标签请求
// @return 应用层创建标签入参
func toCreateTagInput(req *types.CreateTagReq) *tagDto.CreateTagReq {
	return &tagDto.CreateTagReq{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}
}

// toCreateTagRes 将应用层创建标签出参转换为创建标签响应
// @param output 应用层创建标签出参
// @return 创建标签响应
func toCreateTagRes(output *tagDto.CreateTagRes) *types.CreateTagRes {
	res := &types.CreateTagRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.Name = output.Name
	res.Description = output.Description
	res.Color = output.Color
	res.SortId = output.SortId
	return res
}

// toUpdateTagInput 将更新标签请求转换为应用层入参
// @param req 更新标签请求
// @return 应用层更新标签入参
func toUpdateTagInput(req *types.UpdateTagReq) *tagDto.UpdateTagReq {
	return &tagDto.UpdateTagReq{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		SortId:      req.SortId,
	}
}

// toBatchUpdateTagInput 将批量更新标签请求转换为应用层入参
// @param req 批量更新标签请求
// @return 应用层批量更新标签入参
func toBatchUpdateTagInput(req *types.BatchUpdateTagReq) *tagDto.BatchUpdateTagReq {
	tags := make([]*tagDto.BatchUpdateTagItem, 0, len(req.Tags))
	for _, tag := range req.Tags {
		tags = append(tags, &tagDto.BatchUpdateTagItem{
			Id:          tag.Id,
			Name:        tag.Name,
			Description: tag.Description,
			Color:       tag.Color,
			SortId:      tag.SortId,
		})
	}
	return &tagDto.BatchUpdateTagReq{Tags: tags}
}

// toBatchUpdateTagRes 将应用层批量更新标签出参转换为批量更新标签响应
// @param output 应用层批量更新标签出参
// @return 批量更新标签响应
func toBatchUpdateTagRes(output *tagDto.BatchUpdateTagRes) *types.BatchUpdateTagRes {
	return &types.BatchUpdateTagRes{
		UpdatedCount: output.UpdatedCount,
		Tags:         toGetTagResList(output.Tags),
	}
}

// toGetTagPreferenceRes 将应用层获取标签偏好出参转换为获取标签偏好响应
// @param output 应用层获取标签偏好出参
// @return 获取标签偏好响应
func toGetTagPreferenceRes(output *tagDto.GetTagPreferenceRes) *types.GetTagPreferenceRes {
	res := &types.GetTagPreferenceRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.TagId = output.TagId
	res.ViewType = output.ViewType
	res.GetOptions = output.GetOptions
	res.Columns = output.Columns
	return res
}

// toUpdateTagPreferenceInput 将更新标签偏好请求转换为应用层入参
// @param req 更新标签偏好请求
// @return 应用层更新标签偏好入参
func toUpdateTagPreferenceInput(req *types.UpdateTagPreferenceReq) *tagDto.UpdateTagPreferenceReq {
	return &tagDto.UpdateTagPreferenceReq{
		ViewType:   req.ViewType,
		GetOptions: req.GetOptions,
		Columns:    req.Columns,
	}
}

// GetTag 获取单个标签信息接入点
// @code 3000x
func (c *TagController) GetTag(ctx *gin.Context) {
	// 1. 获取标签 ID
	tagId := ctx.Param("tagId")
	if tagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30001,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 获取标签信息
	res, err := c.tagApp.GetTag(ctx.Request.Context(), tagId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30002,
			Message: "获取标签信息失败 - " + err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30000,
		Message: "获取标签信息成功",
		Data:    toGetTagRes(res),
	})
}

// CreateTag 创建标签接入点
// @code 3001x
func (c *TagController) CreateTag(ctx *gin.Context) {
	// 1. 获取请求参数
	createTagReq := &types.CreateTagReq{}
	err := ctx.ShouldBindJSON(createTagReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30011,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 2. 创建标签
	res, err := c.tagApp.CreateTag(ctx.Request.Context(), toCreateTagInput(createTagReq))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30012,
			Message: "创建标签失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30010,
		Message: "创建标签成功",
		Data:    toCreateTagRes(res),
	})
}

// UpdateTag 更新标签接入点
// @code 3002x
func (c *TagController) UpdateTag(ctx *gin.Context) {
	// 1. 获取标签 ID
	tagId := ctx.Param("tagId")
	if tagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30021,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 获取请求参数
	var updateTagReq types.UpdateTagReq
	if err := ctx.ShouldBindJSON(&updateTagReq); err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30022,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 更新标签
	err := c.tagApp.UpdateTag(ctx.Request.Context(), tagId, toUpdateTagInput(&updateTagReq))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30023,
			Message: "更新标签失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30020,
		Message: "更新标签成功",
		Data:    tagId,
	})
}

// DeleteTag 删除标签接入点
// @code 3003x
func (c *TagController) DeleteTag(ctx *gin.Context) {
	// 1. 获取标签 ID
	tagId := ctx.Param("tagId")
	if tagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30031,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 删除标签
	err := c.tagApp.DeleteTag(ctx.Request.Context(), tagId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30032,
			Message: "删除标签失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30030,
		Message: "删除标签成功",
		Data:    tagId,
	})
}

// ListTag 获取标签列表接入点
// @code 3004x
func (c *TagController) ListTag(ctx *gin.Context) {
	// 1. 获取标签 ID列表
	tagIdString := ctx.Query("tagIds")
	if tagIdString == "" {
		// 1. 获取标签列表
		res, err := c.tagApp.ListTag(ctx.Request.Context())
		if err != nil {
			Failure(ctx, types.ResponseData{
				Code:    30041,
				Message: "获取标签列表失败",
				Error:   err.Error(),
			})
			return
		}
		// 2. 返回结果
		Success(ctx, types.ResponseData{
			Code:    30040,
			Message: "获取标签列表成功",
			Data:    toGetTagResList(res),
		})
	} else {
		// 1. 转换标签 ID列表为字符串列表
		tagIds := strings.Split(tagIdString, ",")
		if len(tagIds) == 0 {
			Failure(ctx, types.ResponseData{
				Code:    30042,
				Message: "标签 ID 无效",
			})
			return
		}
		// 2. 获取标签列表
		res, err := c.tagApp.ListTagByIds(ctx.Request.Context(), tagIds)
		if err != nil {
			Failure(ctx, types.ResponseData{
				Code:    30043,
				Message: "获取标签列表失败",
				Error:   err.Error(),
			})
			return
		}
		// 3. 返回结果
		Success(ctx, types.ResponseData{
			Code:    30040,
			Message: "获取标签列表成功",
			Data:    toGetTagResList(res),
		})
	}
}

// GetTagPreference 获取标签偏好接入点
// @code 3005x
func (c *TagController) GetTagPreference(ctx *gin.Context) {
	// 1. 获取标签 ID
	tagId := ctx.Param("tagId")
	if tagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30051,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 获取标签偏好
	res, err := c.tagApp.GetTagPreference(ctx.Request.Context(), tagId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30052,
			Message: "获取标签偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30050,
		Message: "获取标签偏好成功",
		Data:    toGetTagPreferenceRes(res),
	})
}

// UpdateTagPreference 更新标签偏好接入点
// @code 3006x
func (c *TagController) UpdateTagPreference(ctx *gin.Context) {
	// 1. 获取标签 ID
	tagId := ctx.Param("tagId")
	if tagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30061,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 绑定参数
	var req types.UpdateTagPreferenceReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30062,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 更新标签偏好
	err = c.tagApp.UpdateTagPreference(
		ctx.Request.Context(),
		tagId,
		toUpdateTagPreferenceInput(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30063,
			Message: "更新标签偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30060,
		Message: "更新标签偏好成功",
		Data:    tagId,
	})
}

// BatchUpdateTags 批量更新标签接入点
// @code 3007x
func (c *TagController) BatchUpdateTags(ctx *gin.Context) {
	var req types.BatchUpdateTagReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30071,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}
	res, err := c.tagApp.BatchUpdateTags(ctx.Request.Context(), toBatchUpdateTagInput(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30072,
			Message: "批量更新标签失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    30070,
		Message: "批量更新标签成功",
		Data:    toBatchUpdateTagRes(res),
	})
}
