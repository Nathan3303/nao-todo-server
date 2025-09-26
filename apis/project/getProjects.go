package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"

	"github.com/gin-gonic/gin"
)

type GetProjectsHandlerV1DTO struct {
	Name         string `form:"name"`
	Desscription string `form:"description"`
	IsArchived   *bool  `form:"isArchived"`
	IsDeleted    *bool  `form:"isDeleted"`
}

func GetProjectsHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetProjectsHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20051,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取筛选值
	if err := ctx.ShouldBindQuery(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20052,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 构建查询
	var tx = core.DB.Preload("Preference").Unscoped().Where("user_id = ?", userId)
	if dto.Name != "" {
		tx = tx.Where("name like ?", "%"+dto.Name+"%")
	}
	if dto.Desscription != "" {
		tx = tx.Where("description like ?", "%"+dto.Desscription+"%")
	}
	if dto.IsArchived != nil {
		if *dto.IsArchived {
			tx = tx.Where("archived_at IS NOT NULL")
		} else {
			tx = tx.Where("archived_at IS NULL")
		}
	}
	if dto.IsDeleted != nil {
		if *dto.IsDeleted {
			tx = tx.Where("deleted_at IS NOT NULL")
		} else {
			tx = tx.Where("deleted_at IS NULL")
		}
	}

	// 执行查询
	var projectsRaw []models.Project
	tx.Find(&projectsRaw)

	// 转换雪花 ID
	projects := make([]ProjectResponseDTO, len(projectsRaw))
	for i, project := range projectsRaw {
		projects[i] = ToProjectResponse(&project)
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20050,
		Message: "成功获取项目列表",
		Data:    projects,
	})
}
