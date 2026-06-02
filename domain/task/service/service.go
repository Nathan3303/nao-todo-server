package service

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
)

type TaskDomain interface {
	// === Task ===
	GetById(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)
	Create(ctx context.Context, userId int64, vo *valueobjects.CreateTask) (*entities.Task, error)
	Update(ctx context.Context, userId int64, taskId int64, vo *valueobjects.UpdateTask) error
	Delete(ctx context.Context, userId int64, taskId int64) error
	Restore(ctx context.Context, userId int64, taskId int64) error
	Copy(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)
	List(ctx context.Context, userId int64, query *valueobjects.QueryTask, pagination *valueobjects.Pagination) ([]*entities.Task, *valueobjects.Pagination, error)
	Snooze(ctx context.Context, userId int64, taskId int64, durationMinutes int) (string, error)
	ProcessReminders(ctx context.Context) ([]*entities.Task, error)

	// === CheckItem ===
	GetCheckItemById(ctx context.Context, userId, checkItemId int64) (*entities.CheckItem, error)
	CreateCheckItem(ctx context.Context, userId int64, vo *valueobjects.CreateCheckItem) (*entities.CheckItem, error)
	UpdateCheckItem(ctx context.Context, userId, checkItemId int64, vo *valueobjects.UpdateCheckItem) error
	DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error
	ListCheckItems(ctx context.Context, userId, taskId int64) ([]*entities.CheckItem, error)
	BatchUpdateCheckItems(ctx context.Context, userId int64, vos []*valueobjects.BatchUpdateCheckItem) ([]*entities.CheckItem, error)

	// === Comment ===
	GetCommentById(ctx context.Context, userId, commentId int64) (*entities.Comment, error)
	CreateComment(ctx context.Context, userId int64, vo *valueobjects.CreateComment) (*entities.Comment, error)
	UpdateComment(ctx context.Context, userId, commentId int64, vo *valueobjects.UpdateComment) error
	DeleteComment(ctx context.Context, userId, commentId int64) error
	ListComments(ctx context.Context, userId, taskId int64) ([]*entities.Comment, error)
	SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error
}

type TaskDomainImpl struct {
	taskRepo repositories.Task
}
