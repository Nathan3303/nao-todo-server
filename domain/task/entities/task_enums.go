package entities

// TaskState 任务状态
type TaskState uint8

const (
	TaskStatePending    TaskState = 1 // 待办
	TaskStateInProgress TaskState = 2 // 进行中
	TaskStateCompleted  TaskState = 3 // 已完成
)

// TaskPriority 任务优先级
type TaskPriority uint8

const (
	TaskPriorityNone   TaskPriority = 0
	TaskPriorityLow    TaskPriority = 1
	TaskPriorityMedium TaskPriority = 2
	TaskPriorityHigh   TaskPriority = 3
	TaskPriorityUrgent TaskPriority = 4
)

// RemindRepeat 提醒重复类型
type RemindRepeat uint8

const (
	RemindRepeatNone    RemindRepeat = 0 // 不重复
	RemindRepeatDaily   RemindRepeat = 1 // 每天
	RemindRepeatWeekly  RemindRepeat = 2 // 每周
	RemindRepeatMonthly RemindRepeat = 3 // 每月
)

// ParseTaskState 字符串转换任务状态
// @param s 状态字符串
// @return TaskState 任务状态
// @return bool 是否匹配成功
func ParseTaskState(s string) (TaskState, bool) {
	switch s {
	case "todo":
		return TaskStatePending, true
	case "in-progress":
		return TaskStateInProgress, true
	case "done":
		return TaskStateCompleted, true
	}
	return 0, false
}

// String 任务状态转换字符串
// @return string 状态字符串，未匹配返回空字符串
func (s TaskState) String() string {
	switch s {
	case TaskStatePending:
		return "todo"
	case TaskStateInProgress:
		return "in-progress"
	case TaskStateCompleted:
		return "done"
	}
	return ""
}

// ParseTaskPriority 字符串转换任务优先级
// @param s 优先级字符串
// @return TaskPriority 任务优先级
// @return bool 是否匹配成功
func ParseTaskPriority(s string) (TaskPriority, bool) {
	switch s {
	case "low":
		return TaskPriorityLow, true
	case "medium":
		return TaskPriorityMedium, true
	case "high":
		return TaskPriorityHigh, true
	case "urgent":
		return TaskPriorityUrgent, true
	}
	return 0, false
}

// String 任务优先级转换字符串
// @return string 优先级字符串，无优先级与未匹配均返回空字符串
func (p TaskPriority) String() string {
	switch p {
	case TaskPriorityLow:
		return "low"
	case TaskPriorityMedium:
		return "medium"
	case TaskPriorityHigh:
		return "high"
	case TaskPriorityUrgent:
		return "urgent"
	}
	return ""
}

// ParseRemindRepeat 字符串转换提醒重复类型
// @param s 提醒重复类型字符串
// @return RemindRepeat 提醒重复类型
// @return bool 是否匹配成功
func ParseRemindRepeat(s string) (RemindRepeat, bool) {
	switch s {
	case "none":
		return RemindRepeatNone, true
	case "daily":
		return RemindRepeatDaily, true
	case "weekly":
		return RemindRepeatWeekly, true
	case "monthly":
		return RemindRepeatMonthly, true
	}
	return 0, false
}

// String 提醒重复类型转换字符串
// @return string 提醒重复类型字符串，未匹配返回空字符串
func (r RemindRepeat) String() string {
	switch r {
	case RemindRepeatNone:
		return "none"
	case RemindRepeatDaily:
		return "daily"
	case RemindRepeatWeekly:
		return "weekly"
	case RemindRepeatMonthly:
		return "monthly"
	}
	return ""
}

// WeekdaysToBitmask 星期数组转换位掩码
// 星期取值 0..6（0 为周日），超出范围的取值跳过
// @param weekdays 星期数组
// @return uint8 位掩码
func WeekdaysToBitmask(weekdays []uint8) uint8 {
	var mask uint8
	for _, d := range weekdays {
		if d > 6 {
			continue
		}
		mask |= 1 << d
	}
	return mask
}

// BitmaskToWeekdays 位掩码转换星期数组
// 按星期 0..6 升序输出，保证结果确定
// @param mask 位掩码
// @return []uint8 星期数组
func BitmaskToWeekdays(mask uint8) []uint8 {
	weekdays := []uint8{}
	for d := uint8(0); d <= 6; d++ {
		if mask&(1<<d) != 0 {
			weekdays = append(weekdays, d)
		}
	}
	return weekdays
}
