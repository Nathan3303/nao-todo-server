package service

import (
	"context"
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/repositories"
)

type EventDomain interface {
	GetById(ctx context.Context, userId int64, eventId int64) (*entities.Event, error)
	Create(ctx context.Context, userId int64, createEntity *entities.Event) (*entities.Event, error)
	Update(ctx context.Context, userId int64, eventId int64, updateEntity *entities.Event) error
	Delete(ctx context.Context, userId int64, eventId int64) error
	List(ctx context.Context, userId int64, taskId int64) ([]*entities.Event, error)
}

type EventDomainImpl struct {
	eventRepo repositories.Event
}
