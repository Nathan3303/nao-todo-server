package types

// --- Pomodoro ---

// PomodoroRes 常用番茄工作响应
type PomodoroRes struct {
	ResBase
	Type          uint8  `json:"type"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Duration      uint16 `json:"duration"`
	ArchivedAt    string `json:"archivedAt"`
	TotalDuration uint64 `json:"totalDuration"`
}

// CreatePomodoroReq 创建常用番茄工作请求
type CreatePomodoroReq struct {
	Type        uint8  `json:"type" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Duration    uint16 `json:"duration" binding:"required"`
}

// CreatePomodoroRes 创建常用番茄工作响应
type CreatePomodoroRes PomodoroRes

// UpdatePomodoroReq 更新常用番茄工作请求
// 指针字段表示 PATCH 语义：nil 时不更新该字段
type UpdatePomodoroReq struct {
	Type        *uint8  `json:"type"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Duration    *uint16 `json:"duration"`
	ArchivedAt  string  `json:"archivedAt"`
}

// GetPomodoroReq 获取常用番茄工作请求
type GetPomodoroReq struct {
	Id string `json:"id" binding:"required"`
}

// ListPomodoroReq 获取常用番茄工作列表请求
type ListPomodoroReq struct {
	Type       uint8  `json:"type"`
	Name       string `json:"name"`
	IsArchived bool   `json:"isArchived"`
	// ArchivedAt string `json:"archivedAt"`
	ListReqBase
}

// ListPomodoroRes 获取常用番茄工作列表响应
type ListPomodoroRes []PomodoroRes

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
	PomodoroId string `form:"pomodoroId"`
	SessionId  string `form:"sessionId"` // Like uuidv4
	StartTime  string `form:"startTime"`
	EndTime    string `form:"endTime"`
	TaskId     string `form:"taskId"`
	TaskName   string `form:"taskName"`
	Type       uint8  `form:"type"`
	ListReqBase
}

// ListPomodoroRecordRes 获取番茄工作记录列表响应
type ListPomodoroRecordRes []GetPomodoroRecordRes
