package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateProjectHandlerV1DTO struct {
	ProjectIdRaw string
	ProjectId    int64
	Name         string `json:"name"`
	Description  string `json:"description"`
}

func UpdateProjectHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateProjectHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20021,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取项目 ID
	if dto.ProjectIdRaw = ctx.Param("projectId"); dto.ProjectIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20022,
			Message: "项目 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20023,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 构建更新结构体
	var projectCond models.Project
	projectCond.UpdatedAt = time.Time(time.Now())
	if dto.Name != "" {
		projectCond.Name = dto.Name
	}
	if dto.Description != "" {
		projectCond.Description = dto.Description
	}

	// 更新记录
	result := core.DB.Where("id = ? and user_id = ?", dto.ProjectId, userId).UpdateColumns(&projectCond)
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20025,
			Message: "清单更新失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20020,
		Message: "清单更新成功",
		Data:    dto.ProjectIdRaw,
	})
}
