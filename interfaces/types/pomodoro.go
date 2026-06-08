package types

// CreatePomodoroReq 创建番茄请求
type CreatePomodoroReq struct {
	SessionId   string `json:"sessionId" binding:"required"` // Like uuidv4
	Type        uint8  `json:"type"`
	TaskId      string `json:"taskId" binding:"required"`
	TaskName    string `json:"taskName" binding:"required"`
	Description string `json:"description"`
	StartAt     string `json:"startAt" binding:"required"`
	EndAt       string `json:"endAt" binding:"required"`
	Duration    int    `json:"duration" binding:"required"`
	Note        string `json:"note"`
}

// CreatePomodoroRes 创建番茄响应
type CreatePomodoroRes struct {
	Id          string `json:"id"`
	SessionId   string `json:"sessionId"` // Like uuidv4
	Type        uint8  `json:"type"`
	TaskId      string `json:"taskId"`
	TaskName    string `json:"taskName"`
	Description string `json:"description"`
	StartAt     string `json:"startAt"`
	EndAt       string `json:"endAt"`
	Duration    int    `json:"duration"`
	Note        string `json:"note"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// GetPomodoroReq 获取番茄请求
type GetPomodoroReq struct {
	Id string `json:"id" binding:"required"`
}

// GetPomodoroRes 获取番茄响应
type GetPomodoroRes struct {
	Id          string `json:"id"`
	SessionId   string `json:"sessionId"` // Like uuidv4
	Type        uint8  `json:"type"`
	TaskId      string `json:"taskId"`
	TaskName    string `json:"taskName"`
	Description string `json:"description"`
	StartAt     string `json:"startAt"`
	EndAt       string `json:"endAt"`
	Duration    int    `json:"duration"`
	Note        string `json:"note"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	DeletedAt   string `json:"deletedAt"`
}

// ListPomodoroReq 获取番茄列表请求
type ListPomodoroReq struct {
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
