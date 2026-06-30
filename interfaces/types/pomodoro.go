package types

// --- Pomodoro ---

// ...

// --- PomodoroRecord ---

// CreatePomodoroRecordReq 创建番茄工作记录请求
type CreatePomodoroRecordReq struct {
	SessionId   string `json:"sessionId" binding:"required"` // Like uuidv4
	Type        uint8  `json:"type"`
	TaskId      string `json:"taskId" binding:"required"`
	TaskName    string `json:"taskName" binding:"required"`
	Description string `json:"description"`
	StartAt     string `json:"startAt" binding:"required"`
	EndAt       string `json:"endAt" binding:"required"`
	Duration    uint16 `json:"duration" binding:"required"`
	Note        string `json:"note"`
}

// CreatePomodoroRecordRes 创建番茄工作记录响应
type CreatePomodoroRecordRes struct {
	ResBase
	SessionId   string `json:"sessionId"` // Like uuidv4
	Type        uint8  `json:"type"`
	TaskId      string `json:"taskId"`
	TaskName    string `json:"taskName"`
	Description string `json:"description"`
	StartAt     string `json:"startAt"`
	EndAt       string `json:"endAt"`
	Duration    uint16 `json:"duration"`
	Note        string `json:"note"`
}

// GetPomodoroRecordReq 获取番茄工作记录请求
type GetPomodoroRecordReq struct {
	Id string `json:"id" binding:"required"`
}

// GetPomodoroRecordRes 获取番茄工作记录响应
type GetPomodoroRecordRes CreatePomodoroRecordRes

// ListPomodoroRecordReq 获取番茄工作记录列表请求
type ListPomodoroRecordReq struct {
	SessionId string `form:"sessionId"` // Like uuidv4
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	TaskId    string `form:"taskId"`
	TaskName  string `form:"taskName"`
	Type      uint8  `form:"type"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
	Sort      string `form:"sort"` // Like: 'field:order'
}

// ListPomodoroRecordRes 获取番茄工作记录列表响应
type ListPomodoroRecordRes []GetPomodoroRecordRes
