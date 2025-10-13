package apis

import (
	"math"
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type GetTodosHandlerV1DTO struct {
	ProjectIdRaw string `form:"projectId"`
	ProjectId    int64
	TagIdRaw     string `form:"tagId"`
	TagId        int64
	Name         string `form:"name"`
	Description  string `form:"description"`
	StateRaw     string `form:"state"`
	PriorityRaw  string `form:"priority"`
	StartAtRaw   string `form:"startAt"`
	StartAt      *time.Time
	EndAtRaw     string `form:"endAt"`
	EndAt        *time.Time
	DeletedRaw   string `form:"isDeleted"`
	ArchivedRaw  string `form:"isArchived"`
	FavoritedRaw string `form:"isFavorited"`
	GivenUpRaw   string `form:"isGivenUp"`
	PageRaw      *int   `form:"page"`
	Page         int
	LimitRaw     *int `form:"limit"`
	Limit        int
	RelativeDate *string `form:"relativeDate"`
	Sort         string  `form:"sort"`
}

func GetTodosHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto GetTodosHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40051,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性值
	if err := ctx.ShouldBindQuery(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40052,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 校验时间信息
	if dto.EndAtRaw != "" {
		dto.EndAt, _ = ParseDateString("2006-01-02 15:04:05", dto.EndAtRaw)
	}
	if dto.StartAtRaw != "" {
		dto.StartAt, _ = ParseDateString("2006-01-02 15:04:05", dto.StartAtRaw)
	}
	if dto.EndAt != nil && dto.StartAt != nil && dto.StartAt.After(*dto.EndAt) {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40053,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 校验分页信息
	if dto.PageRaw == nil || *dto.PageRaw <= 0 {
		dto.Page = 1
	} else {
		dto.Page = *dto.PageRaw
	}
	if dto.LimitRaw == nil || *dto.LimitRaw <= 0 {
		dto.Limit = 10
	} else {
		dto.Limit = *dto.LimitRaw
	}

	// 构建查询
	var tx = core.DB.Model(&models.Todo{}).Where("user_id = ?", userId)
	{
		if dto.ProjectIdRaw != "" {
			if dto.ProjectIdRaw == "inbox" {
				dto.ProjectId = userId.(int64)
			} else {
				dto.ProjectId, _ = strconv.ParseInt(dto.ProjectIdRaw, 10, 64)
			}
			tx = tx.Where("project_id = ?", dto.ProjectId)
		}
		if dto.TagIdRaw != "" {
			tx = tx.Where("tags LIKE ?", "%"+dto.TagIdRaw+"%")
		}
		if dto.Name != "" {
			tx = tx.Where("name LIKE ?", "%"+dto.Name+"%")
		}
		if dto.Description != "" {
			tx = tx.Where("description LIKE ?", "%"+dto.Description+"%")
		}
		if dto.StateRaw != "" {
			var statesArray []int8
			for state := range strings.SplitSeq(dto.StateRaw, ",") {
				statesArray = append(statesArray, TodoStateMap[state])
			}
			tx.Where("state in ?", statesArray)
		}
		if dto.PriorityRaw != "" {
			var statesArray []int8
			for state := range strings.SplitSeq(dto.PriorityRaw, ",") {
				statesArray = append(statesArray, TodoPriorityMap[state])
			}
			tx.Where("priority in ?", statesArray)
		}
		if dto.StartAt != nil {
			tx = tx.Where("start_at >= ?", dto.StartAt)
		}
		if dto.EndAt != nil {
			tx = tx.Where("end_at <= ?", dto.EndAt)
		}
		if dto.DeletedRaw == "true" {
			tx = tx.Where("deleted_at IS NOT NULL")
		}
		if dto.ArchivedRaw == "true" {
			tx = tx.Where("archived_at IS NOT NULL")
		}
		if dto.FavoritedRaw == "true" {
			tx = tx.Where("favorited_at IS NOT NULL")
		}
		if dto.GivenUpRaw == "true" {
			tx = tx.Where("given_up_at IS NOT NULL")
		}
		if dto.RelativeDate != nil {
			ParseRelativeDateToUpdateCond(tx, *dto.RelativeDate)
		}
		if dto.Sort != "" {
			ParseSortStringToQueryCond(tx, dto.Sort)
		}
	}

	// 查询总数
	var total int64
	if countErr := tx.Count(&total).Error; countErr != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40054,
			Message: "查询失败",
			Data:    nil,
		})
		return
	}

	// 执行查询
	var todosRaw []models.Todo
	if findErr := tx.Offset((dto.Page - 1) * dto.Limit).Limit(dto.Limit).Find(&todosRaw).Error; findErr != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40055,
			Message: "查询失败",
			Data:    nil,
		})
		return
	}

	// 转换查询结果
	var todos []models.TodoResponse
	for _, todo := range todosRaw {
		todos = append(todos, ToTodoResponse(&todo))
	}

	// 返回结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40050,
		Message: "待办列表获取成功",
		Data:    todos,
		Pagination: &apis.Pagination{
			Total:   int(total),
			Page:    dto.Page,
			Limit:   dto.Limit,
			MaxPage: int(math.Ceil(float64(total) / float64(dto.Limit))),
		},
	})
}
