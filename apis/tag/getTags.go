package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"

	"github.com/gin-gonic/gin"
)

type GetTagsHandlerV1DTO struct {
	Name         string
	Desscription string
	CreatedAt    string
	ArchivedAt   string
	Color        string
}

func GetTagsHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetTagsHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30051,
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
	dto.Color = ctx.Query("color")

	// 构建查询条件
	var tx = core.DB.Preload("Preference").Where("user_id = ?", userId)
	if dto.Name != "" {
		tx = tx.Where("name like ?", "%"+dto.Name+"%")
	}
	if dto.Desscription != "" {
		tx = tx.Where("description like ?", "%"+dto.Desscription+"%")
	}
	if dto.Color != "" {
		tx = tx.Where("color = ?", dto.Color)
	}

	// 执行查询
	var tags []modules.Tag
	tx.Find(&tags)

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    30050,
		Message: "成功获取标签列表",
		Data:    tags,
	})
}
