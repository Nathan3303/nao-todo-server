package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"

	"github.com/gin-gonic/gin"
)

type CreateProjectHandlerV1DTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func CreateProjectHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto CreateProjectHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20011,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20012,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 检查属性值
	if dto.Name == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20013,
			Message: "清单名称不能为空",
			Data:    nil,
		})
		return
	}

	// 创建记录
	var projectPreference = &models.ProjectPreference{
		UserId:     userId.(int64),
		ViewType:   "table",
		GetOptions: "{\"limit\": 20}",
		Columns:    "{\"createdAt\":false,\"updatedAt\":false,\"description\":true,\"state\":true,\"priority\":true,\"tags\":true,\"startAt\":false,\"endAt\":true}",
	}
	var projectRaw = &models.Project{
		Name:        dto.Name,
		Description: dto.Description,
		UserId:      userId.(int64),
		Preference:  projectPreference,
	}
	result := core.DB.Create(&projectRaw)
	if result.Error != nil || projectRaw.ID == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    20014,
			Message: "创建清单失败",
			Data:    nil,
		})
		return
	}

	// 返回成功结果
	apis.Success(ctx, apis.ResponseData{
		Code:    20010,
		Message: "创建清单成功",
		Data:    ToProjectResponse(projectRaw),
	})
}
