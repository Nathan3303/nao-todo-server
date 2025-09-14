package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetProjectHandlerV1DTO struct {
	ProjectIdRaw string
	ProjectId    int64
}

func GetProjectHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetProjectHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20001,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if dto.ProjectIdRaw = ctx.Query("projectId"); dto.ProjectIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20002,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}
	dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)

	// 获取项目
	var project models.Project
	core.DB.Preload("Preference").Where("id = ? and user_id = ?", dto.ProjectId, userId).First(&project)
	if project.ID == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20003,
			Message: "项目不存在",
			Data:    nil,
		})
		return
	}

	// 返回项目信息
	apis.Success(ctx, apis.ResponseData{
		Code:    200,
		Message: "获取项目成功",
		Data:    project,
	})
}
