package sse

import (
	"context"

	"naotodoserver/domain/types"
)

// notificationPublisherImpl NotificationPublisher 实现
// 内部通过 SSE Hub 将领域事件推送到订阅连接
type notificationPublisherImpl struct {
	hub *Hub
}

// NewNotificationPublisher 创建 NotificationPublisher 实例
// @return types.NotificationPublisher 通知发布端口实例
func NewNotificationPublisher() types.NotificationPublisher {
	return &notificationPublisherImpl{hub: GetHub()}
}

// PublishReminder 推送任务提醒事件
// @param ctx 上下文
// @param userId 接收方用户 ID
// @param event 提醒事件载荷
// @return error 推送错误
func (p *notificationPublisherImpl) PublishReminder(
	_ context.Context,
	userId int64,
	event types.ReminderEvent,
) error {
	p.hub.Publish(userId, ReminderEvent(event))
	return nil
}
