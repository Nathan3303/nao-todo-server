package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RestoreProjectHandlerV1DTO struct {
	ProjectId    int64
	ProjectIdRaw string
}

func RestoreProjectHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto RestoreProjectHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20041,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取项目 ID
	if dto.ProjectIdRaw = ctx.Param("projectId"); dto.ProjectIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20042,
			Message: "项目 ID 不能为空",
			Data:    nil,
		})
		return
	}
	dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)

	// 恢复记录
	var project modules.Project
	core.DB.Unscoped().Where("id = ? and user_id = ?", dto.ProjectId, userId).First(&project)
	if project.DeletedAt.Valid { // 确认已被软删除
		project.DeletedAt = gorm.DeletedAt{} // 手动清空 DeletedAt
		core.DB.Unscoped().Save(&project)    // 保存，恢复记录
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20040,
		Message: "恢复成功",
		Data:    project.ID,
	})
}
