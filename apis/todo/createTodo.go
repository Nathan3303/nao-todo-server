package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateTodoHandlerV1DTO struct {
	ProjectIdRaw string `json:"projectId" binding:"required"`
	ProjectId    int64
	Name         string  `json:"name" binding:"required"`
	Description  string  `json:"description"`
	State        string  `json:"state" binding:"required"`
	Priority     string  `json:"priority" binding:"required"`
	StartAtRaw   *string `json:"startAt"`
	StartAt      *time.Time
	EndAtRaw     *string `json:"endAt"`
	EndAt        *time.Time
	Tags         []string `json:"tags"`
}

func CreateTodoHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto CreateTodoHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40011,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40012,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 转换清单 ID
	if dto.ProjectIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40013,
			Message: "清单 ID 无效",
			Data:    nil,
		})
		return
	}
	dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)

	// 校验基本信息
	if dto.Name == "" || dto.State == "" || dto.Priority == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40014,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 校验时间信息
	var errOfEndAtParsing, errOfStartAtParsing error
	if dto.EndAtRaw != nil {
		dto.EndAt, errOfEndAtParsing = ParseDateString(time.RFC3339, *dto.EndAtRaw)
		if errOfEndAtParsing != nil {
			apis.Failure(ctx, apis.ResponseData{
				Code:    40015,
				Message: "参数错误",
				Data:    errOfEndAtParsing.Error(),
			})
			return
		}
	}
	if dto.StartAtRaw != nil {
		dto.StartAt, errOfStartAtParsing = ParseDateString(time.RFC3339, *dto.StartAtRaw)
		if errOfStartAtParsing != nil {
			apis.Failure(ctx, apis.ResponseData{
				Code:    40015,
				Message: "参数错误",
				Data:    errOfStartAtParsing.Error(),
			})
			return
		}
		if dto.StartAt.After(*dto.EndAt) {
			apis.Failure(ctx, apis.ResponseData{
				Code:    40016,
				Message: "参数错误",
				Data:    nil,
			})
			return
		}
	}

	// 创建记录
	var todo = &models.Todo{
		UserId:      userId.(int64),
		ProjectId:   dto.ProjectId,
		Name:        dto.Name,
		Description: dto.Description,
		State:       TodoStateMap[dto.State],
		Priority:    TodoPriorityMap[dto.Priority],
		StartAt:     dto.StartAt,
		EndAt:       dto.EndAt,
		Tags:        dto.Tags,
	}
	result := core.DB.Create(&todo)

	// 判断结果
	if todo.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40017,
			Message: "创建任务失败",
			Data:    result.Error.Error(),
		})
		return
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40010,
		Message: "创建任务成功",
		Data:    ToTodoResponse(todo),
	})
}
