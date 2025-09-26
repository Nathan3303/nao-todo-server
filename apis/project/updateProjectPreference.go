package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateProjectPreferenceDTO struct {
	ProjectIdRaw    string
	ProjectId       int64
	Columns         string `json:"columns"`
	GetTodosOptions string `json:"getTodosOptions"`
	ViewType        string `json:"viewType"`
}

func UpdateProjectPreferenceHandlerV1(ctx *gin.Context) {
	// 定义 请求 DTO
	var dto UpdateProjectPreferenceDTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20061,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取项目 ID
	if dto.ProjectIdRaw = ctx.Param("projectId"); dto.ProjectIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20062,
			Message: "清单 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)

	// 获取参数
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20063,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 构建更新结构体
	var projectPreferenceCond models.ProjectPreference
	projectPreferenceCond.UpdatedAt = time.Time(time.Now())
	if dto.Columns != "" {
		projectPreferenceCond.Columns = dto.Columns
	}
	if dto.GetTodosOptions != "" {
		projectPreferenceCond.GetOptions = dto.GetTodosOptions
	}
	if dto.ViewType != "" {
		projectPreferenceCond.ViewType = dto.ViewType
	}

	// 更新记录
	result := core.DB.Where("project_id = ? and user_id = ?", dto.ProjectId, userId).UpdateColumns(&projectPreferenceCond)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20064,
			Message: "清单偏好更新失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20060,
		Message: "清单偏好更新成功",
		Data:    userId,
	})
}
