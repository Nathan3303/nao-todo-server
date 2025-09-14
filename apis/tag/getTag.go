package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GetTagHandlerV1DTO struct {
	TagIdRaw string
	TagId    int64
}

func GetTagHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetTagHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30001,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if dto.TagIdRaw = ctx.Param("tagId"); dto.TagIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30002,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}
	dto.TagId, _ = strconv.ParseInt(dto.TagIdRaw, 10, 64)

	// 获取标签
	var tag models.Tag
	core.DB.Preload("Preference").Where("id = ? and user_id = ?", dto.TagId, userId).First(&tag)
	if tag.ID == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30003,
			Message: "标签不存在",
			Data:    nil,
		})
		return
	}

	// 返回标签信息
	apis.Success(ctx, apis.ResponseData{
		Code:    30000,
		Message: "获取标签成功",
		Data:    tag,
	})
}
