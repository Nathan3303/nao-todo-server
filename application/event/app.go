package event

import (
	"context"
	"naotodoserver/domain/event/service"
	"naotodoserver/interfaces/types"
	"sync"
)

// 检查事项应用接口
type EventApp interface {
	// 获取检查事项详情
	GetEventById(ctx context.Context, eventId string) (*types.GetEventRes, error)

	// 创建检查事项
	CreateEvent(ctx context.Context, req *types.CreateEventReq) (*types.CreateEventRes, error)

	// 更新检查事项
	UpdateEvent(
		ctx context.Context,
		eventId string,
		req *types.UpdateEventReq,
	) (*types.UpdateEventRes, error)

	// 删除检查事项
	DeleteEvent(ctx context.Context, eventId string) (*types.DeleteEventRes, error)

	// 获取检查事项列表
	ListEvent(ctx context.Context, taskId string) (types.ListEventRes, error)

	// 排序检查事项
	ResortEvents(ctx context.Context, req *types.ResortEventsReq) (*types.ResortEventsRes, error)
}

// 检查事项应用实现
type EventAppImpl struct {
	eventDomain service.EventDomain
}

// 检查事项应用实例
var (
	App  EventApp
	once sync.Once
)
