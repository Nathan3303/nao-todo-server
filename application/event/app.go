package event

import (
	"context"
	"naotodoserver/domain/event/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type EventApp interface {
	GetEventById(ctx context.Context, eventId string) (*types.EventRes, error)
	CreateEvent(ctx context.Context, req types.CreateEventReq) (*types.EventRes, error)
	UpdateEvent(
		ctx context.Context,
		eventId string,
		req types.UpdateEventReq,
	) (*types.UpdateEventRes, error)
	DeleteEvent(ctx context.Context, eventId string) (*types.DeleteEventRes, error)
	ListEvent(ctx context.Context, taskId string) (types.ListEventRes, error)
	ResortEvents(ctx context.Context, req types.ResortEventsReq) (*types.ResortEventsRes, error)
}

type EventAppImpl struct {
	eventDomain service.EventDomain
}

var (
	App  EventApp
	once sync.Once
)
