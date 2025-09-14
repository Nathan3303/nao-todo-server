package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"

	"github.com/gin-gonic/gin"
)

type GetProjectsHandlerV1DTO struct {
	Name         string
	Desscription string
	CreatedAt    string
	ArchivedAt   string
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
	dto.Name = ctx.Query("name")
	dto.Desscription = ctx.Query("description")
	// dto.CreatedAt = ctx.Query("created_at")
	// dto.ArchivedAt = ctx.Query("archived_at")

	// 执行查询
	var projects []modules.Project
	core.DB.Preload("Preference").Where("user_id = ?", userId).Find(
		&projects,
		"name like ? and description like ?",
		"%"+dto.Name+"%",
		"%"+dto.Desscription+"%",
	)

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20050,
		Message: "成功获取项目列表",
		Data:    projects,
	})
}
