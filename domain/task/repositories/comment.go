package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
)

// TaskComment 任务评论仓库接口
type TaskComment interface {
	// GetCommentById 获取单个任务评论信息
	GetCommentById(ctx context.Context, userId, commentId int64) (*entities.TaskComment, error)

	// CreateComment 创建任务评论
	CreateComment(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskComment,
	) (*entities.TaskComment, error)

	// UpdateComment 更新任务评论
	UpdateComment(
		ctx context.Context,
		userId int64,
		commentId int64,
		vo *valueobjects.UpdateTaskComment,
	) error

	// DeleteComment 删除任务评论
	DeleteComment(ctx context.Context, userId, commentId int64) error

	// ListComments 获取任务评论列表
	ListComments(ctx context.Context, userId, taskId int64) ([]*entities.TaskComment, error)

	// SyncCommentUserProfile 同步任务评论用户信息
	SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}
