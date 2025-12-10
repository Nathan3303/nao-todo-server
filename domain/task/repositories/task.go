package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/vo"
)

type Task interface {
	GetById(ctx context.Context, whereEntity *entities.Task) (*entities.Task, error)
	Create(ctx context.Context, createEntity *entities.Task) (*entities.Task, error)
	Update(ctx context.Context, whereEntity *entities.Task, updateEntity *entities.Task) error
	Delete(ctx context.Context, whereEntity *entities.Task) error
	Restore(ctx context.Context, whereEntity *entities.Task) error
	List(
		ctx context.Context,
		whereEntity *entities.Task,
		pagination *vo.Pagination,
	) ([]*entities.Task, *vo.Pagination, error)
}
