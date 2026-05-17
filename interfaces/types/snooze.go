package types

// SnoozeTaskReq 稍后提醒请求
type SnoozeTaskReq struct {
	DurationMinutes int `json:"durationMinutes" binding:"required,min=1,max=1440"`
}

// SnoozeTaskRes 稍后提醒响应
type SnoozeTaskRes struct {
	RemindAt string `json:"remindAt"`
}
