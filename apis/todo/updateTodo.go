package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UpdateTodoHandlerV1DTO struct {
	TodoIdRaw    string
	TodoId       int64
	ProjectIdRaw string `json:"projectId"`
	ProjectId    int64
	Name         string `json:"name"`
	Description  string `json:"description"`
	StateRaw     string `json:"state"`
	State        int8
	PriorityRaw  string `json:"priority"`
	Priority     int8
	StartAtRaw   string `json:"startAt"`
	StartAt      *time.Time
	EndAtRaw     string `json:"endAt"`
	EndAt        *time.Time
	Tags         []string `json:"tags"`
	FavoritedRaw *bool    `json:"isFavorited"`
}

func UpdateTodoHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto UpdateTodoHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40021,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Param("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40022,
			Message: "待办 ID 不能为空",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 获取属性值
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40023,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 校验时间信息
	if dto.EndAtRaw != "" {
		dto.EndAt, _ = ParseDateString(time.RFC3339, dto.EndAtRaw)
	}
	if dto.StartAtRaw != "" {
		dto.StartAt, _ = ParseDateString(time.RFC3339, dto.StartAtRaw)
	}
	if dto.EndAt != nil && dto.StartAt != nil && dto.StartAt.After(*dto.EndAt) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40024,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 构建更新结构体
	var (
		todoCond models.Todo
		vErr     error
	)
	todoCond.UpdatedAt = time.Time(time.Now())
	{
		if dto.ProjectIdRaw != "" {
			dto.ProjectId, vErr = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)
			if vErr == nil {
				todoCond.ProjectId = dto.ProjectId
			}
		}
		if dto.Name != "" {
			todoCond.Name = dto.Name
		}
		if dto.Description != "" {
			todoCond.Description = dto.Description
		}
		if dto.StateRaw != "" {
			todoCond.State = TodoStateMap[dto.StateRaw]
		}
		if dto.PriorityRaw != "" {
			todoCond.Priority = TodoPriorityMap[dto.PriorityRaw]
		}
		if dto.StartAt != nil {
			todoCond.StartAt = dto.StartAt
		}
		if dto.EndAt != nil {
			todoCond.EndAt = dto.EndAt
		}
		if dto.Tags != nil {
			todoCond.Tags = dto.Tags
		}
		if dto.FavoritedRaw != nil && *dto.FavoritedRaw {
			var time = time.Now()
			todoCond.FavoritedAt = &time
		}
	}

	// 执行更新
	result := core.DB.Where("id = ? and user_id = ?", dto.TodoId, userId).UpdateColumns(&todoCond)

	// 判断结果
	if result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40025,
			Message: "更新待办失败",
			Data:    nil,
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40020,
		Message: "更新待办成功",
		Data:    dto.TodoIdRaw,
	})
}
