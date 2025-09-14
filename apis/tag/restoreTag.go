package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/modules"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RestoreTagHandlerV1DTO struct {
	TagId    int64
	TagIdRaw string
}

func RestoreTagHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto RestoreTagHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30041,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取标签 ID
	if dto.TagIdRaw = ctx.Param("tagId"); dto.TagIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30042,
			Message: "标签 ID 不能为空",
			Data:    nil,
		})
		return
	}
	dto.TagId, _ = strconv.ParseInt(dto.TagIdRaw, 10, 64)

	// 恢复记录
	var tag modules.Tag
	core.DB.Unscoped().Where("id = ? and user_id = ?", dto.TagId, userId).First(&tag)
	if tag.DeletedAt.Valid { // 确认已被软删除
		tag.DeletedAt = gorm.DeletedAt{} // 手动清空 DeletedAt
		core.DB.Unscoped().Save(&tag)    // 保存，恢复记录
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    30040,
		Message: "标签恢复成功",
		Data:    tag.ID,
	})
}
