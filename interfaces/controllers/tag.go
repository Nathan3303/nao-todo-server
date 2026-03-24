package controllers

import (
	"naotodoserver/application/tag"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

/*
 * Get tag handler
 * 获取单个标签信息处理函数（30000）
 */
func GetTagHandler(ctx *gin.Context) {
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
	res, err := tag.App.GetTag(ctx.Request.Context(), tagId)
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
		Data:    res,
	})
}

/*
 * Create tag handler
 * 创建标签处理函数（30010）
 */
func CreateTagHandler(ctx *gin.Context) {
	// 1. 获取请求参数
	req := &types.CreateTagReq{}
	err := ctx.ShouldBindJSON(req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30011,
			Message: "请求参数错误 - " + err.Error(),
		})
		return
	}
	// 2. 创建标签
	res, err := tag.App.CreateTag(ctx.Request.Context(), req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30012,
			Message: "创建标签失败 - " + err.Error(),
		})
		return
	}
	// 3. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30010,
		Message: "创建标签成功",
		Data:    res,
	})
}

/*
 * Update tag handler
 * 更新标签处理函数（30020）
 */
func UpdateTagHandler(ctx *gin.Context) {
	udpateTagReq := &types.UpdateTagReq{}
	// 1. 获取标签 ID
	udpateTagReq.TagId = ctx.Param("tagId")
	if udpateTagReq.TagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30021,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 获取请求参数
	if err := ctx.ShouldBindJSON(udpateTagReq); err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30022,
			Message: "请求参数错误 - " + err.Error(),
		})
		return
	}
	// 3. 更新标签
	res, err := tag.App.UpdateTag(ctx.Request.Context(), udpateTagReq)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30023,
			Message: "更新标签失败 - " + err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30020,
		Message: "更新标签成功",
		Data:    res,
	})
}

/*
 * Delete tag handler
 * 删除标签处理函数（30030）
 */
func DeleteTagHandler(ctx *gin.Context) {
	// 1. 获取标签 ID
	tagId := ctx.Param("tagId")
	if tagId == "" {
		Failure(ctx, types.ResponseData{
			Code:    30031,
			Message: "标签 ID 不能为空",
		})
		return
	}
	// 2. 更新标签
	res, err := tag.App.DeleteTag(ctx.Request.Context(), tagId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30032,
			Message: "删除标签失败 - " + err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30030,
		Message: "删除标签成功",
		Data:    res,
	})
}

/*
 * List tag handler
 * 获取标签列表处理函数（30040）
 */
func ListTagHandler(ctx *gin.Context) {
	// 1. 更新标签
	res, err := tag.App.ListTag(ctx.Request.Context())
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30041,
			Message: "获取标签列表失败 - " + err.Error(),
		})
		return
	}
	// 2. 返回结果
	Success(ctx, types.ResponseData{
		Code:    30040,
		Message: "获取标签列表成功",
		Data:    res,
	})
}

/*
 * Get tag preference handler
 * 获取标签偏好处理函数（30050）
 */
func GetTagPreferenceHandler(ctx *gin.Context) {
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
	res, err := tag.App.GetTagPreference(ctx.Request.Context(), tagId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30052,
			Message: "获取标签偏好失败 - " + err.Error(),
		})
		return
	}
	// 3. 响应结果
	Success(ctx, types.ResponseData{
		Code:    30050,
		Message: "获取标签偏好成功",
		Data:    res,
	})
}

/*
 * Update tag preference handler
 * 更新标签偏好处理函数（30060）
 */
func UpdateTagPreferenceHandler(ctx *gin.Context) {
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
	res, err := tag.App.UpdateTagPreference(ctx.Request.Context(), tagId, &req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    30063,
			Message: "更新标签偏好失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 响应结果
	Success(ctx, types.ResponseData{
		Code:    30060,
		Message: "更新标签偏好成功",
		Data:    res,
	})
}
