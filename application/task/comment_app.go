package task

import (
	"context"
	"naotodoserver/interfaces/types"
)

// TaskCommentApp 任务评论应用接口
type TaskCommentApp interface {
	GetTaskCommentById(ctx context.Context, commentId string) (*types.TaskCommentRes, error)
	CreateTaskComment(
		ctx context.Context, req *types.CreateTaskCommentReq,
	) (*types.TaskCommentRes, error)
	UpdateTaskComment(ctx context.Context, commentId string, req *types.UpdateTaskCommentReq) error
	DeleteTaskComment(ctx context.Context, commentId string) error
	ListTaskComments(ctx context.Context, taskId string) ([]*types.TaskCommentRes, error)
	SyncTaskCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}
