package service

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/vo"
)

type TaskDomain interface {
	GetById(ctx context.Context, userId int64, taskId int64) (*entities.Task, error)
	Create(ctx context.Context, userId int64, createEntity *entities.Task) (*entities.Task, error)
	Update(
		ctx context.Context,
		userId int64,
		taskId int64,
		updateEntity *entities.Task,
	) error
	Delete(ctx context.Context, userId int64, taskId int64) error
	Restore(ctx context.Context, userId int64, taskId int64) error
	List(
		ctx context.Context,
		userId int64,
		query *vo.TaskQuery,
		pagination *vo.Pagination,
	) ([]*entities.Task, *vo.Pagination, error)
}

type TaskDomainImpl struct {
	taskRepo repositories.Task
}
