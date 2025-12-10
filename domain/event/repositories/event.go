package repositories

import (
	"context"
	"naotodoserver/domain/event/entities"
)

type Event interface {
	GetById(ctx context.Context, userId int64, eventId int64) (*entities.Event, error)
	Create(ctx context.Context, createEntity *entities.Event) (*entities.Event, error)
	Update(ctx context.Context, whereEntity *entities.Event, updateEntity *entities.Event) error
	Delete(ctx context.Context, whereEntity *entities.Event) error
	Get(ctx context.Context, whereEntity *entities.Event) ([]*entities.Event, error)
}
