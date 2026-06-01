package event

import (
	"context"
	"naotodoserver/domain/checkitem/service"
	"naotodoserver/interfaces/types"
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
	) error

	// 删除检查事项
	DeleteEvent(ctx context.Context, eventId string) error

	// 获取检查事项列表
	ListEvent(ctx context.Context, taskId string) (types.ListEventRes, error)

	// 排序检查事项
	ResortEvents(ctx context.Context, req *types.ResortEventsReq) (*types.ResortEventsRes, error)

	// 批量更新检查事项
	BatchUpdateEvents(ctx context.Context, req *types.BatchUpdateEventReq) (*types.BatchUpdateEventRes, error)
}

// 检查事项应用实现
type EventAppImpl struct {
	eventDomain service.EventDomain
}

// 检查事项应用实例
