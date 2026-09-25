package types

import "context"

// CountEventType 计数事件类型（领域统计属性联动，ADR 2026-09-12 §4.3）
type CountEventType string

const (
	// CountEventCheckItemChanged 检查项计数变化（E1）：TaskId 所属任务，Delta ±1
	CountEventCheckItemChanged CountEventType = "check_item_count_changed"
	// CountEventCommentChanged 评论计数变化（E2）：TaskId 所属任务，Delta ±1
	CountEventCommentChanged CountEventType = "comment_count_changed"
	// CountEventSubTaskChanged 直接子任务计数变化（E3/E6）：TaskId 被计数父任务，Delta ±1
	CountEventSubTaskChanged CountEventType = "sub_task_count_changed"
	// CountEventTaskCountChanged 项目任务计数变化（E4）：ProjectId 项目，Delta ±1
	CountEventTaskCountChanged CountEventType = "task_count_changed"
	// CountEventTaskMoved 任务移动（E5）：OldProjectId/ProjectId = 旧/新项目
	CountEventTaskMoved CountEventType = "task_moved"
	// CountEventTaskParentChanged 任务换父/脱离（E6）：OldParentTaskId/ParentTaskId = 旧/新父任务
	CountEventTaskParentChanged CountEventType = "task_parent_changed"
	// CountEventTaskCountRecounted 项目任务计数批量重算（E7）：ProjectId 项目，Delta 忽略
	CountEventTaskCountRecounted CountEventType = "task_count_recounted"
)

// CountEvent 计数事件载荷
// 领域层事件载体，无 JSON tag（仿 ReminderEvent）；字段按事件类型取义，见各常量注释。
type CountEvent struct {
	Type            CountEventType
	UserId          int64
	TaskId          int64 // E1/E2 所属任务；E3 被计数父任务
	ParentTaskId    int64 // E6 新父任务
	OldParentTaskId int64 // E6 旧父任务
	ProjectId       int64 // E4/E7 项目；E5 新项目
	OldProjectId    int64 // E5 旧项目
	Delta           int8  // ±1（E7 重算忽略）
}

// CountEventPublisher 计数事件发布端口
// 由 infrastructure/events 实现（进程内同步内存总线），供 application 层写路径调用；
// application 层只依赖该接口，具体实现由 infrastructure 注入（仿 NotificationPublisher 先例）。
type CountEventPublisher interface {
	// PublishCountEvent 发布计数事件（同步分发）
	// 事件在写事务内、提交前同步分发；返回错误 ⇒ 调用方应回滚整个事务（计数与主写永不分离）
	PublishCountEvent(ctx context.Context, event CountEvent) error
}
