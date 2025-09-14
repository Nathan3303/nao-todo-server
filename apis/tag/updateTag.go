package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateTagHandlerV1DTO struct {
	TagIdRaw    string
	TagId       int64
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func UpdateTagHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateTagHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30021,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取标签 ID
	if dto.TagIdRaw = ctx.Param("tagId"); dto.TagIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30022,
			Message: "标签 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.TagId, _ = strconv.ParseInt(dto.TagIdRaw, 10, 64)

	// 获取参数
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30023,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 检查属性值
	if dto.Name == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30024,
			Message: "标签名称不能为空",
			Data:    nil,
		})
		return
	}

	// 更新记录
	var tag modules.Tag
	tag.UpdatedAt = time.Time(time.Now())
	tag.Name = dto.Name
	tag.Description = dto.Description
	tag.Color = dto.Color
	result := core.DB.Where("id = ? and user_id = ?", dto.TagId, userId).UpdateColumns(&tag).Preload("Preference").First(&tag)
	if result.RowsAffected == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30025,
			Message: "标签更新失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    30020,
		Message: "标签更新成功",
		Data:    tag.ID,
	})
}
