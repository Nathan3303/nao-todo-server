package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteTagHandlerV1DTO struct {
	TagIdRaw        string
	TagId           int64
	isHardDeleteRaw string
	isHardDelete    bool
}

func DeleteTagHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto DeleteTagHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30031,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取标签 ID
	if dto.TagIdRaw = ctx.Param("tagId"); dto.TagIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30032,
			Message: "标签 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.TagId, _ = strconv.ParseInt(dto.TagIdRaw, 10, 64)

	// 获取是否硬删除
	dto.isHardDeleteRaw = ctx.Query("hard")
	dto.isHardDelete = dto.isHardDeleteRaw == "true"

	// 执行删除
	var result *gorm.DB
	if dto.isHardDelete {
		result = core.DB.Unscoped().Where("id = ? and user_id = ?", dto.TagId, userId).Delete(&models.Tag{})
	} else {
		result = core.DB.Where("id = ? and user_id = ?", dto.TagId, userId).Delete(&models.Tag{})
	}

	// 判断删除结果
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    30033,
			Message: "标签删除失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    30030,
		Message: "标签删除成功",
		Data:    dto.TagId,
	})
}
