package task

import (
	"context"
	"time"

	"naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
)

// TaskCommentApp 任务评论应用接口
type TaskCommentApp interface {
	GetTaskCommentById(
		ctx context.Context, userId int64, commentId string,
	) (*dto.TaskCommentRes, error)
	CreateTaskComment(
		ctx context.Context, userId int64, req *dto.CreateTaskCommentReq,
	) (*dto.TaskCommentRes, domaintypes.UpsertResult, error)
	UpdateTaskComment(
		ctx context.Context, userId int64, commentId string, req *dto.UpdateTaskCommentReq,
	) error
	// DeleteTaskComment 删除评论（含 OCC 裁决，T642 / A4）：返回库中当前版本与裁决
	DeleteTaskComment(
		ctx context.Context, userId int64, commentId string, baseUpdatedAt time.Time,
	) (*dto.TaskCommentRes, domaintypes.DeleteDecision, error)
	// GetTaskCommentSyncRow 读取单行评论的同步形状（含墓碑），供 serverCurrent 使用
	GetTaskCommentSyncRow(
		ctx context.Context, userId int64, commentId string,
	) (*dto.TaskCommentRes, error)
	ListTaskComments(
		ctx context.Context, userId int64, taskId string,
	) ([]*dto.TaskCommentRes, error)
	ListTaskCommentSync(
		ctx context.Context, userId int64, updatedAt, cursorId string, limit int,
	) ([]*dto.TaskCommentRes, error)
	SyncTaskCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}
