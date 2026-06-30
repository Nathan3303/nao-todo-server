package service

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
)

// TaskDomain 任务域接口
type TaskDomain interface {
	// --- Task ---
	GetById(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)
	Create(ctx context.Context, userId int64, vo *valueobjects.CreateTask) (*entities.Task, error)
	Update(ctx context.Context, userId int64, taskId int64, vo *valueobjects.UpdateTask) error
	Delete(ctx context.Context, userId int64, taskId int64) error
	Restore(ctx context.Context, userId int64, taskId int64) error
	Copy(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)
	List(
		ctx context.Context,
		userId int64,
		query *valueobjects.QueryTask,
		pagination *valueobjects.Pagination,
	) ([]*entities.Task, *valueobjects.Pagination, error)

	// --- CheckItem ---
	GetCheckItemById(
		ctx context.Context,
		userId int64,
		checkItemId int64,
	) (*entities.TaskCheckItem, error)
	CreateCheckItem(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskCheckItem,
	) (*entities.TaskCheckItem, error)
	UpdateCheckItem(
		ctx context.Context,
		userId int64,
		checkItemId int64,
		vo *valueobjects.UpdateTaskCheckItem,
	) error
	DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error
	ListCheckItems(ctx context.Context, userId, taskId int64) ([]*entities.TaskCheckItem, error)
	BatchUpdateCheckItems(
		ctx context.Context,
		userId int64,
		vos []*valueobjects.BatchUpdateTaskCheckItem,
	) ([]*entities.TaskCheckItem, error)

	// --- Comment ---
	GetCommentById(ctx context.Context, userId, commentId int64) (*entities.TaskComment, error)
	CreateComment(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskComment,
	) (*entities.TaskComment, error)
	UpdateComment(
		ctx context.Context,
		userId int64,
		commentId int64,
		vo *valueobjects.UpdateTaskComment,
	) error
	DeleteComment(ctx context.Context, userId, commentId int64) error
	ListComments(ctx context.Context, userId, taskId int64) ([]*entities.TaskComment, error)
	SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error

	// --- Snooze ---
	Snooze(ctx context.Context, userId int64, taskId int64, durationMinutes int) (string, error)
	ProcessReminders(ctx context.Context) ([]*entities.Task, error)
}

// TaskDomainImpl 任务域实现
type TaskDomainImpl struct {
	taskRepo repositories.Task
}
