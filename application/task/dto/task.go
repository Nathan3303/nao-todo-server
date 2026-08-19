// Package dto 定义任务应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// --- Task ---

// GetTaskRes 获取任务出参
type GetTaskRes struct {
	Id             string   // 任务 ID
	CreatedAt      string   // 创建时间
	UpdatedAt      string   // 更新时间
	DeletedAt      string   // 删除时间
	ParentTaskId   string   // 父任务 ID
	Name           string   // 任务名称
	Description    string   // 任务描述
	State          string   // 任务状态
	Priority       string   // 任务优先级
	StartAt        string   // 开始时间
	EndAt          string   // 结束时间
	Tags           []string // 标签列表
	ProjectId      string   // 项目 ID
	ArchivedAt     string   // 归档时间
	StarMarkAt     string   // 星标时间
	GivenUpAt      string   // 放弃时间
	RemindAt       string   // 提醒时间
	RemindRepeat   string   // 提醒重复规则
	RemindTime     string   // 提醒具体时间
	RemindWeekdays []uint8  // 提醒星期
	SortId         uint16   // 排序值
}

// CreateTaskReq 创建任务入参
type CreateTaskReq struct {
	ParentTaskId   string   // 父任务 ID
	Name           string   // 任务名称
	Description    string   // 任务描述
	State          string   // 任务状态
	Priority       string   // 任务优先级
	StartAt        string   // 开始时间
	EndAt          string   // 结束时间
	ProjectId      string   // 项目 ID
	Tags           []string // 标签列表
	RemindAt       string   // 提醒时间
	RemindRepeat   string   // 提醒重复规则
	RemindTime     string   // 提醒具体时间
	RemindWeekdays []uint8  // 提醒星期
	Id             *string  // 同步元数据：客户端预置 id
	CreatedAt      *string  // 同步元数据：客户端预置 createdAt
	UpdatedAt      *string  // 同步元数据：客户端预置 updatedAt
	DeletedAt      *string  // 同步元数据：客户端预置 deletedAt（推送本地墓碑时携带）
}

// UpdateTaskReq 更新任务入参
type UpdateTaskReq struct {
	ParentTaskId   *string  // 父任务 ID
	Name           *string  // 任务名称
	Description    *string  // 任务描述
	State          *string  // 任务状态
	Priority       *string  // 任务优先级
	StartAt        *string  // 开始时间
	EndAt          *string  // 结束时间
	ProjectId      *string  // 项目 ID
	Tags           []string // 标签列表
	ArchivedAt     *string  // 归档时间
	StarMarkAt     *string  // 星标时间
	GivenUpAt      *string  // 放弃时间
	RemindAt       *string  // 提醒时间
	RemindRepeat   *string  // 提醒重复规则
	RemindTime     *string  // 提醒具体时间
	RemindWeekdays []uint8  // 提醒星期
	SortId         *uint16  // 排序值
	UpdatedAt      *string  // 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
}

// ListTaskReq 列表任务入参
type ListTaskReq struct {
	ParentTaskId string // 父任务 ID
	ProjectId    string // 项目 ID
	TagId        string // 标签 ID
	Name         string // 任务名称
	Description  string // 任务描述
	State        string // 任务状态
	Priority     string // 任务优先级
	StartAt      string // 开始时间
	EndAt        string // 结束时间
	DeletedAt    string // 删除时间
	ArchivedAt   string // 归档时间
	StarMarkAt   string // 星标时间
	GivenUpAt    string // 放弃时间
	IsDeleted    bool   // 是否已删除
	IsArchived   bool   // 是否已归档
	IsStarMarked bool   // 是否已星标
	IsGivenUp    bool   // 是否已放弃
	RelativeDate string // 相对日期
	UpdatedAt    string // 增量同步游标（RFC3339，updated_at > 该值）
	CursorId     string // keyset 游标辅助：与 UpdatedAt 组合 (updated_at, id) > (cursor, cursorId)
	Page         int    // 当前页数
	Limit        int    // 每页条数
	Sort         string // 排序字段
}

// ListTaskRes 列表任务出参
type ListTaskRes []*GetTaskRes

// Pagination 分页出参
type Pagination struct {
	Total   int64 // 总记录数
	Page    int   // 当前页数
	Limit   int   // 每页条数
	MaxPage int   // 最大页数
}

// SnoozeTaskReq 稍后提醒入参
type SnoozeTaskReq struct {
	DurationMinutes int // 稍后提醒分钟数
}

// SnoozeTaskRes 稍后提醒出参
type SnoozeTaskRes struct {
	RemindAt string // 新的提醒时间
}

// --- TaskCheckItem ---

// GetTaskCheckItemRes 获取任务检查项出参
type GetTaskCheckItemRes struct {
	Id          string // 检查项 ID
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
	DeletedAt   string // 删除时间
	TaskId      string // 所属任务 ID
	Name        string // 检查项名称
	Description string // 检查项描述
	IsDone      bool   // 是否已完成
	SortId      uint16 // 排序值
}

// CreateTaskCheckItemReq 创建任务检查项入参
type CreateTaskCheckItemReq struct {
	TaskId      string // 所属任务 ID
	Name        string // 检查项名称
	Description string // 检查项描述
	Id          *string // 同步元数据：客户端预置 id
	CreatedAt   *string // 同步元数据：客户端预置 createdAt
	UpdatedAt   *string // 同步元数据：客户端预置 updatedAt
}

// CreateTaskCheckItemRes 创建任务检查项出参
type CreateTaskCheckItemRes struct {
	Id          string // 检查项 ID
	CreatedAt   string // 创建时间
	UpdatedAt   string // 更新时间
	DeletedAt   string // 删除时间
	TaskId      string // 所属任务 ID
	Name        string // 检查项名称
	Description string // 检查项描述
	IsDone      bool   // 是否已完成
	SortId      uint16 // 排序值
}

// UpdateTaskCheckItemReq 更新任务检查项入参
type UpdateTaskCheckItemReq struct {
	Name        *string // 检查项名称
	Description *string // 检查项描述
	IsDone      *bool   // 是否已完成
	SortId      *uint16 // 排序值
	UpdatedAt   *string // 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
}

// ListTaskCheckItemRes 列表任务检查项出参
type ListTaskCheckItemRes []*GetTaskCheckItemRes

// BatchUpdateTaskCheckItemEvent 批量更新任务检查项事件项
type BatchUpdateTaskCheckItemEvent struct {
	Id          string  // 检查项 ID
	Name        *string // 检查项名称
	Description *string // 检查项描述
	IsDone      *bool   // 是否已完成
	SortId      *uint16 // 排序值
}

// BatchUpdateTaskCheckItemReq 批量更新任务检查项入参
type BatchUpdateTaskCheckItemReq struct {
	Events []*BatchUpdateTaskCheckItemEvent // 事件项列表
}

// BatchUpdateTaskCheckItemRes 批量更新任务检查项出参
type BatchUpdateTaskCheckItemRes struct {
	UpdatedCount int64                  // 更新数量
	Events       []*GetTaskCheckItemRes // 更新后的检查项列表
}

// --- TaskComment ---

// TaskCommentRes 任务评论出参
type TaskCommentRes struct {
	Id          string   // 评论 ID
	CreatedAt   string   // 创建时间
	UpdatedAt   string   // 更新时间
	DeletedAt   string   // 删除时间
	TaskId      string   // 所属任务 ID
	Content     string   // 评论内容
	Attachments []string // 附件列表
	IsTopUp     bool     // 是否置顶
	Nickname    string   // 用户昵称
	Avatar      string   // 用户头像
}

// CreateTaskCommentReq 创建任务评论入参
type CreateTaskCommentReq struct {
	TaskId    string // 所属任务 ID
	Content   string // 评论内容
	Id        *string // 同步元数据：客户端预置 id
	CreatedAt *string // 同步元数据：客户端预置 createdAt
	UpdatedAt *string // 同步元数据：客户端预置 updatedAt
}

// UpdateTaskCommentReq 更新任务评论入参
type UpdateTaskCommentReq struct {
	Content     *string   // 评论内容
	Attachments *[]string // 附件列表
	IsTopUp     *bool     // 是否置顶
	UpdatedAt   *string   // 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
}
