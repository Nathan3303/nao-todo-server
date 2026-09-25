package types

import "context"

// NotificationPublisher 通知发布端口
// 由 infrastructure/sse 实现,供 application 层调用以推送领域事件
// application 层只依赖该接口,具体实现由 infrastructure 注入
type NotificationPublisher interface {
	// PublishReminder 推送任务提醒事件
	// @param ctx 上下文
	// @param userId 接收方用户 ID
	// @param event 提醒事件载荷
	// @return error 推送错误
	PublishReminder(ctx context.Context, userId int64, event ReminderEvent) error
}

// ReminderEvent 提醒事件载荷
// 领域层事件载体,无 JSON tag,由实现层负责序列化
type ReminderEvent struct {
	Type        string
	TaskId      string
	TaskName    string
	Description string
	RemindAt    string
}
