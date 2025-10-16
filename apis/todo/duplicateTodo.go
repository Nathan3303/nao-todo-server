package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DuplicateTodoHandlerV1DTO struct {
	TodoIdRaw string
	TodoId    int64
}

func DuplicateTodoHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto DuplicateTodoHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40061,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	if dto.TodoIdRaw = ctx.Param("todoId"); dto.TodoIdRaw == "" {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40062,
			Message: "待办 ID 不能为空",
			Data:    nil,
		})
		return
	}
	dto.TodoId, _ = strconv.ParseInt(dto.TodoIdRaw, 10, 64)

	// 获取待办评论
	// var todoComments []models.Comment
	// var tx1 = core.DB.Model(&models.Comment{}).Preload("CommentUser").Where("todo_id = ? and user_id = ?", dto.TodoId, userId)
	// if err := tx1.Find(&todoComments).Error; err != nil {
	// 	apis.Failure(ctx, apis.ResponseData{
	// 		Code:    40063,
	// 		Message: "获取待办评论失败",
	// 		Data:    nil,
	// 	})
	// 	return
	// }

	// 获取待办检查事项
	var todoEvents []models.Event
	var tx2 = core.DB.Model(&models.Event{}).Where("todo_id = ? and user_id = ?", dto.TodoId, userId)
	if err := tx2.Find(&todoEvents).Error; err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40064,
			Message: "获取待办检查事项失败",
			Data:    nil,
		})
		return
	}

	// 获取待办
	var todo models.Todo
	var tx3 = core.DB.Model(&models.Todo{}).Where("id = ? and user_id = ?", dto.TodoId, userId)
	if err := tx3.Find(&todo).Error; err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40065,
			Message: "获取待办失败",
			Data:    nil,
		})
		return
	}

	// 复制待办事务
	var newTodo models.Todo
	err := core.DB.Transaction(func(tx *gorm.DB) error {
		// 复制待办
		newTodo = models.Todo{
			UserId:       userId.(int64),
			ProjectId:    todo.ProjectId,
			ParentTodoId: todo.ParentTodoId,
			Name:         "Copy from " + todo.Name,
			Description:  todo.Description,
			State:        todo.State,
			Priority:     todo.Priority,
			StartAt:      todo.StartAt,
			EndAt:        todo.EndAt,
			ArchivedAt:   todo.ArchivedAt,
			FavoritedAt:  todo.FavoritedAt,
			GivenUpAt:    todo.GivenUpAt,
			Tags:         todo.Tags,
		}
		if err := core.DB.Model(&models.Todo{}).Create(&newTodo).Error; err != nil {
			apis.Failure(ctx, apis.ResponseData{
				Code:    40066,
				Message: "创建待办失败",
				Data:    nil,
			})
			return err
		}

		// 复制待办评论
		// var newComments []models.Comment
		// for _, comment := range todoComments {
		// 	newComment := comment
		// 	newComment.ID = 0
		// 	newComment.TodoId = newTodo.ID
		// 	newComment.CreatedAt = time.Now()
		// 	newComment.CommentUser.ID = 0
		// 	newComments = append(newComments, newComment)
		// }
		// if len(newComments) > 0 {
		// 	if err := core.DB.Model(&models.Comment{}).Create(&newComments).Error; err != nil {
		// 		return err
		// 	}
		// }

		// 复制待办检查事项
		var newEvents []models.Event
		for _, event := range todoEvents {
			newEvent := event
			newEvent.ID = 0
			newEvent.TodoId = newTodo.ID
			newEvent.CreatedAt = time.Now()
			newEvents = append(newEvents, newEvent)
		}
		if len(newEvents) > 0 {
			if err := core.DB.Model(&models.Event{}).Create(&newEvents).Error; err != nil {
				return err
			}
		}

		return nil
	})

	// 处理失败结果
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    40067,
			Message: "复制待办失败",
			Data:    nil,
		})
		return
	}

	// 处理成功结果
	apis.Success(ctx, apis.ResponseData{
		Code:    40060,
		Message: "复制待办成功",
		Data:    ToTodoResponse(&newTodo),
	})
}
