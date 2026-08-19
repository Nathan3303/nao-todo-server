// Package dto 定义专注应用应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// ResBase 基础响应字段
type ResBase struct {
	Id        string // 唯一 ID
	CreatedAt string // 创建时间（RFC3339）
	UpdatedAt string // 更新时间（RFC3339）
	DeletedAt string // 删除时间（RFC3339，空表示未删除）
}

// --- PomodoroRecord ---

// CreatePomodoroRecordReq 创建番茄工作记录入参
type CreatePomodoroRecordReq struct {
	SessionId   string // Like uuidv4
	PomodoroId  string // 关联的常用番茄工作 ID（弱关联，可为空）
	Type        uint8  // 番茄工作类型
	TaskId      string // 关联的任务 ID（弱关联，可为空）
	TaskName    string // 任务名称
	Description string // 描述
	StartAt     string // 开始时间（RFC3339）
	EndAt       string // 结束时间（RFC3339）
	Duration    uint16 // 时长（秒）
	Note        string // 备注
	Id          *string // 同步元数据：客户端预置 id
	CreatedAt   *string // 同步元数据：客户端预置 createdAt
	UpdatedAt   *string // 同步元数据：客户端预置 updatedAt
}

// CreatePomodoroRecordRes 创建番茄工作记录出参
type CreatePomodoroRecordRes struct {
	ResBase
	SessionId   string // Like uuidv4
	PomodoroId  string // 关联的常用番茄工作 ID
	Type        uint8  // 番茄工作类型
	TaskId      string // 关联的任务 ID
	TaskName    string // 任务名称
	Description string // 描述
	StartAt     string // 开始时间（RFC3339）
	EndAt       string // 结束时间（RFC3339）
	Duration    uint16 // 时长（秒）
	Note        string // 备注
}

// GetPomodoroRecordReq 获取番茄工作记录入参
type GetPomodoroRecordReq struct {
	Id string // 番茄工作记录 ID
}

// GetPomodoroRecordRes 获取番茄工作记录出参
type GetPomodoroRecordRes CreatePomodoroRecordRes

// ListPomodoroRecordReq 获取番茄工作记录列表入参
type ListPomodoroRecordReq struct {
	PomodoroId string // 关联的常用番茄工作 ID
	SessionId  string // Like uuidv4
	StartTime  string // 开始时间范围（RFC3339）
	EndTime    string // 结束时间范围（RFC3339）
	TaskId     string // 关联的任务 ID
	TaskName   string // 任务名称
	Type       uint8  // 番茄工作类型
	UpdatedAt  string // 增量同步游标（RFC3339，updated_at > 该值）
	CursorId   string // keyset 游标辅助：与 UpdatedAt 组合 (updated_at, id) > (cursor, cursorId)
	Page       int    // 当前页数
	Limit      int    // 每页条数
	Sort       string // 排序规则
}

// --- Pomodoro ---

// CreatePomodoroReq 创建常用番茄工作入参
type CreatePomodoroReq struct {
	Type        uint8  // 番茄工作类型
	Name        string // 名称
	Description string // 描述
	Duration    uint16  // 时长（秒）
	Id          *string // 同步元数据：客户端预置 id
	CreatedAt   *string // 同步元数据：客户端预置 createdAt
	UpdatedAt   *string // 同步元数据：客户端预置 updatedAt
	DeletedAt   *string // 同步元数据：客户端预置 deletedAt（推送本地墓碑时携带）
}

// PomodoroRes 常用番茄工作出参
type PomodoroRes struct {
	ResBase
	Type          uint8  // 番茄工作类型
	Name          string // 名称
	Description   string // 描述
	Duration      uint16 // 时长（秒）
	ArchivedAt    string // 归档时间（RFC3339，空表示未归档）
	TotalDuration uint64 // 累计时长（秒）
}

// CreatePomodoroRes 创建常用番茄工作出参
type CreatePomodoroRes PomodoroRes

// GetPomodoroReq 获取常用番茄工作入参
type GetPomodoroReq struct {
	Id string // 常用番茄工作 ID
}

// UpdatePomodoroReq 更新常用番茄工作入参
// 指针字段表示 PATCH 语义：nil 时不更新该字段
type UpdatePomodoroReq struct {
	Type        *uint8  // 番茄工作类型
	Name        *string // 名称
	Description *string // 描述
	Duration    *uint16 // 时长（秒）
	ArchivedAt  string  // 归档时间（RFC3339）
	UpdatedAt   *string // 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
}

// ListPomodoroReq 获取常用番茄工作列表入参
type ListPomodoroReq struct {
	Type       uint8  // 番茄工作类型
	Name       string // 名称
	IsArchived bool   // 是否已归档
	UpdatedAt  string // 增量同步游标（RFC3339，updated_at > 该值）
	CursorId   string // keyset 游标辅助：与 UpdatedAt 组合 (updated_at, id) > (cursor, cursorId)
	Page       int    // 当前页数
	Limit      int    // 每页条数
	Sort       string // 排序规则
}

// ListPomodoroRes 获取常用番茄工作列表出参
type ListPomodoroRes []PomodoroRes
