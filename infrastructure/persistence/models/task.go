package models

import (
	"database/sql"
)

// Task 待办任务模型
type Task struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_task_user_id"`

	// 父待办任务 ID
	ParentTaskId int64 `gorm:"null;index:idx_task_parent"`

	// 待办任务名称
	Name string `gorm:"not null;size:256"`

	// 待办任务描述
	Description string `gorm:"null;type:text"`

	// 待办任务状态
	// 1: 待办
	// 2: 进行中
	// 3: 已完成
	State uint8 `gorm:"default:1;type:tinyint(1);index:idx_task_state"`

	// 任务优先级
	// 1: 高
	// 2: 中
	// 3: 低
	Priority uint8 `gorm:"default:1;type:tinyint(1);index:idx_task_priority"`

	// 任务开始时间
	StartAt sql.NullTime `gorm:"null;index:idx_task_start_at"`

	// 任务结束时间
	EndAt sql.NullTime `gorm:"null;index:idx_task_end_at"`

	// 项目 ID
	ProjectId int64 `gorm:"null;index:idx_task_project"`

	// 任务标签
	Tags []string `gorm:"serializer:json;type:json"`

	// 任务归档时间
	ArchivedAt sql.NullTime `gorm:"null;index:idx_task_archived_at"`

	// 任务星标时间
	StarMarkAt sql.NullTime `gorm:"null;index:idx_task_star_mark_at"`

	// 任务放弃时间
	GivenUpAt sql.NullTime `gorm:"null;index:idx_task_given_up_at"`

	// 任务完成时间
	CompletedAt sql.NullTime `gorm:"null;index:idx_task_completed_at"`

	// --- 任务提醒功能相关属性 ---

	// 任务提醒时间
	// 记录任务下次提醒时间
	// 通常通过 RemindTime 和 RemindWeekdays 计算，或者用户手动设置
	RemindAt sql.NullTime `gorm:"null;index:idx_task_remind_at"`

	// 任务提醒重复次数
	RemindRepeat uint8 `gorm:"default:0;type:tinyint(1)"`

	// 任务提醒时间
	// 时间格式为 HH:mm，例如 20:00、21:30、23:59
	RemindTime string `gorm:"null;size:5"`

	// 任务提醒周几
	// 通过位运算表示，例如 0x01 表示周一，0x02 表示周二，0x04 表示周三 ... 0x00 表示不提醒
	RemindWeekdays uint8 `gorm:"default:0;type:tinyint(1)"`
}

// TaskCheckItem 待办任务检查项模型
type TaskCheckItem struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_task_check_item_user_id"`

	// 待办任务 ID
	TaskId int64 `gorm:"not null;index:idx_task_check_item_task_id"`

	// 检查项名称
	Name string `gorm:"not null;size:256"`

	// 检查项描述
	Description string `gorm:"null;size:512"`

	// 是否完成
	IsDone bool `gorm:"index:idx_task_check_item_is_done"`

	// 排序 ID
	// 用于在待办任务中显示检查项的顺序
	// 通常采用大间距的整数，例如 100、200 等，更新排序时取被排序项的 SortId +- 1/2
	// 超出类型范围，则需要重新计算所有检查项的 SortId
	SortId uint16 `gorm:"default:0"`
}

// TaskComment 待办任务评论模型
type TaskComment struct {
	// 基础属性
	ModelBase

	// 用户 ID
	UserId int64 `gorm:"not null;index:idx_task_comment_user_id"`

	// 待办任务 ID
	TaskId int64 `gorm:"not null;index:idx_task_comment_task_id"`

	// 评论内容
	Content string `gorm:"not null;type:text"`

	// 评论附件
	// 由附件 UUID 组成的 JSON 数组
	Attachments []string `gorm:"null;serializer:json;type:json"`

	// 是否置顶
	IsTopUp bool `gorm:"index:idx_task_comment_is_top"`

	// 用户属性快照
	// 当 TaskComment 创建时，从用户表获取用户昵称和头像，存储到 TaskComment 中
	// 若用户更新了昵称或头像，需要更新 TaskComment 中的对应字段

	// 评论用户昵称
	Nickname string `gorm:"not null;size:64"`

	// 评论用户头像
	Avatar string `gorm:"not null;size:256"`
}
