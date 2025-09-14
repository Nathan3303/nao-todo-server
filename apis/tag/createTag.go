package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"

	"github.com/gin-gonic/gin"
)

type CreateTagHandlerV1DTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func CreateTagHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto CreateTagHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30011,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30012,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 检查属性值
	if dto.Name == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30013,
			Message: "标签名称不能为空",
			Data:    nil,
		})
		return
	}

	// 创建记录
	var tagPreference = &modules.TagPreference{
		ViewType:   "table",
		GetOptions: "{}",
		Columns:    "priority,project,description,endAt",
	}
	var tag = &modules.Tag{
		UserId:      userId.(int64),
		Name:        dto.Name,
		Description: dto.Description,
		Color:       dto.Color,
		Preference:  tagPreference,
	}
	core.DB.Create(&tag)
	if tag.ID == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30014,
			Message: "创建标签失败",
			Data:    nil,
		})
		return
	}

	// 返回成功结果
	apis.Success(ctx, apis.ResponseData{
		Code:    30010,
		Message: "创建标签成功",
		Data:    tag,
	})
}
