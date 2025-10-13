package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DeleteProjectHandlerV1DTO struct {
	ProjectIdRaw    string
	ProjectId       int64
	isHardDeleteRaw string
	isHardDelete    bool
}

func DeleteProjectHandlerV1(ctx *gin.Context) {
	// 定义转换对象
	var dto DeleteProjectHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20031,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取项目 ID
	if dto.ProjectIdRaw = ctx.Param("projectId"); dto.ProjectIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20032,
			Message: "项目 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)

	// 获取是否硬删除
	dto.isHardDeleteRaw = ctx.Query("hard")
	dto.isHardDelete = dto.isHardDeleteRaw == "true"

	// 执行删除
	if dto.isHardDelete {
		result := core.DB.Unscoped().Where("id = ? and user_id = ?", dto.ProjectId, userId).Delete(&models.Project{})

		// 判断删除结果
		if result.Error != nil {
			apis.Failure(ctx, apis.ResponseData{
				Code:    20033,
				Message: "项目删除失败",
				Data:    nil,
			})
			return
		}

		// 删除所有数据此项目的待办任务
		result = core.DB.Model(&models.Todo{}).Unscoped().Where("user_id = ? and project_id = ?", userId, dto.ProjectId).Delete(&models.Todo{})
		if result.Error != nil {
			apis.Failure(ctx, apis.ResponseData{
				Code:    20034,
				Message: "项目删除失败",
				Data:    nil,
			})
			return
		}

	} else {
		result := core.DB.Where("id = ? and user_id = ?", dto.ProjectId, userId).Delete(&models.Project{})

		// 判断删除结果
		if result.Error != nil {
			apis.Failure(ctx, apis.ResponseData{
				Code:    20035,
				Message: "项目删除失败",
				Data:    nil,
			})
			return
		}
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20030,
		Message: "项目删除成功",
		Data:    dto.ProjectId,
	})
}
