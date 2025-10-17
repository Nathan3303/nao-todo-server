package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateTagPreferenceDTO struct {
	TagIdRaw        string
	TagId           int64
	Columns         string `json:"columns"`
	GetTodosOptions string `json:"getTodosOptions"`
	ViewType        string `json:"viewType"`
}

func UpdateTagPreferenceHandlerV1(ctx *gin.Context) {
	// 定义 请求 DTO
	var dto UpdateTagPreferenceDTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30061,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取标签 ID
	if dto.TagIdRaw = ctx.Param("tagId"); dto.TagIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30062,
			Message: "标签 ID 无效",
			Data:    nil,
		})
		return
	}

	// 获取参数
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30063,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 构建更新结构体
	var tagPreferenceCond models.TagPreference
	tagPreferenceCond.UpdatedAt = time.Now()
	if dto.Columns != "" {
		tagPreferenceCond.Columns = dto.Columns
	}
	if dto.GetTodosOptions != "" {
		tagPreferenceCond.GetOptions = dto.GetTodosOptions
	}
	if dto.ViewType != "" {
		tagPreferenceCond.ViewType = dto.ViewType
	}

	// 更新记录
	result := core.DB.Model(&models.Tag{}).Where("id = ? and user_id = ?", dto.TagId, userId).UpdateColumns(&tagPreferenceCond)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30064,
			Message: "标签偏好更新失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    30060,
		Message: "标签偏好更新成功",
		Data:    userId,
	})
}
