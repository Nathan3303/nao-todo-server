package apis

import (
	"naotodoserver/apis"
	"naotodoserver/core"
	"naotodoserver/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateCommentHandlerV1DTO struct {
	TodoIdRaw string `json:"todoId" binding:"required"`
	TodoId    int64
	Content   string `json:"content" binding:"required"`
	IsTopUp   bool   `json:"isTopUp"`
}

func CreateCommentHandlerV1(ctx *gin.Context) {
	// 定义 DTO
	var dto CreateCommentHandlerV1DTO

	// 获取用户 ID
	userId, exist := ctx.Get("userId")
	if !exist {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60011,
			Message: "用户凭证无效",
			Data:    nil,
		})
		return
	}

	// 获取属性
	if err := ctx.ShouldBind(&dto); err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60012,
			Message: "参数错误",
			Data:    nil,
		})
		return
	}

	// 获取待办 ID
	var err error
	dto.TodoId, err = strconv.ParseInt(dto.TodoIdRaw, 10, 64)
	if err != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60013,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	// 获取用户信息
	var user models.User
	result := core.DB.Where("id = ?", userId).First(&user)
	if result.Error != nil || user.ID == 0 {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60014,
			Message: "评论用户失效",
			Data:    nil,
		})
		return
	}

	// 创建记录
	var commentUser = &models.CommentUser{
		Avatar:   user.Avatar,
		Nickname: user.Nickname,
	}
	var commentRaw = &models.Comment{
		UserId:      userId.(int64),
		TodoId:      dto.TodoId,
		Content:     dto.Content,
		CommentUser: commentUser,
	}
	result = core.DB.Create(&commentRaw)
	if commentRaw.ID == 0 || result.Error != nil {
		apis.Failure(ctx, apis.ResponseData{
			Code:    60014,
			Message: "评论创建失败",
			Data:    nil,
		})
		return
	}

	// 返回成功结果
	apis.Success(ctx, apis.ResponseData{
		Code:    60010,
		Message: "评论创建成功",
		Data:    ToCommentResponse(commentRaw),
	})
}
