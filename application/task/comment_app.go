package task

import (
	"context"
	"naotodoserver/application/task/dto"
)

// TaskCommentApp 任务评论应用接口
type TaskCommentApp interface {
	GetTaskCommentById(ctx context.Context, commentId string) (*dto.TaskCommentRes, error)
	CreateTaskComment(
		ctx context.Context, req *dto.CreateTaskCommentReq,
	) (*dto.TaskCommentRes, error)
	UpdateTaskComment(ctx context.Context, commentId string, req *dto.UpdateTaskCommentReq) error
	DeleteTaskComment(ctx context.Context, commentId string) error
	ListTaskComments(ctx context.Context, taskId string) ([]*dto.TaskCommentRes, error)
	SyncTaskCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}
