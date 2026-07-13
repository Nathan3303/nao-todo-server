package entities

// TaskState 任务状态
type TaskState uint8

const (
	TaskStatePending    TaskState = 0 // 待办
	TaskStateInProgress TaskState = 1 // 进行中
	TaskStateCompleted  TaskState = 2 // 已完成
	TaskStateGivenUp    TaskState = 3 // 已放弃
	TaskStateArchived   TaskState = 4 // 已归档
)

// TaskPriority 任务优先级
type TaskPriority uint8

const (
	TaskPriorityNone    TaskPriority = 0
	TaskPriorityLow     TaskPriority = 1
	TaskPriorityMedium  TaskPriority = 2
	TaskPriorityHigh    TaskPriority = 3
	TaskPriorityUrgent  TaskPriority = 4
)
