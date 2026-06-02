package task

import (
	"context"
	"naotodoserver/domain/task/service"
	"naotodoserver/interfaces/types"
)

type TaskApp interface {
	// === Task ===
	GetTaskById(ctx context.Context, taskId string) (*types.GetTaskRes, error)
	CreateTask(ctx context.Context, req *types.CreateTaskReq) (*types.GetTaskRes, error)
	UpdateTask(ctx context.Context, taskId string, req *types.UpdateTaskReq) error
	DeleteTask(ctx context.Context, taskId string) error
	RestoreTask(ctx context.Context, taskId string) error
	CopyTask(ctx context.Context, taskId string) (*types.GetTaskRes, error)
	ListTask(ctx context.Context, req *types.ListTaskReq) (types.ListTaskRes, *types.Pagination, error)
	SnoozeTask(ctx context.Context, taskId string, req *types.SnoozeTaskReq) (*types.SnoozeTaskRes, error)
	ProcessReminders(ctx context.Context) error

	// === CheckItem ===
	GetCheckItemById(ctx context.Context, checkItemId string) (*types.GetCheckItemRes, error)
	CreateCheckItem(ctx context.Context, req *types.CreateCheckItemReq) (*types.CreateCheckItemRes, error)
	UpdateCheckItem(ctx context.Context, checkItemId string, req *types.UpdateCheckItemReq) error
	DeleteCheckItem(ctx context.Context, checkItemId string) error
	ListCheckItems(ctx context.Context, taskId string) (types.ListCheckItemRes, error)
	BatchUpdateCheckItems(ctx context.Context, req *types.BatchUpdateCheckItemReq) (*types.BatchUpdateCheckItemRes, error)

	// === Comment ===
	GetCommentById(ctx context.Context, commentId string) (*types.CommentRes, error)
	CreateComment(ctx context.Context, req *types.CreateCommentReq) (*types.CommentRes, error)
	UpdateComment(ctx context.Context, commentId string, req *types.UpdateCommentReq) error
	DeleteComment(ctx context.Context, commentId string) error
	ListComments(ctx context.Context, taskId string) ([]*types.CommentRes, error)
	SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}

type TaskAppImpl struct {
	taskDomain service.TaskDomain
}
