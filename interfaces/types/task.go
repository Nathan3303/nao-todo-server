package types

// GetTaskRes 获取任务响应
type GetTaskRes struct {
	ResBase
	ParentTaskId   string   `json:"parentTaskId"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	State          string   `json:"state"`
	Priority       string   `json:"priority"`
	StartAt        string   `json:"startAt"`
	EndAt          string   `json:"endAt"`
	Tags           []string `json:"tags"`
	ProjectId      string   `json:"projectId"`
	ArchivedAt     string   `json:"archivedAt"`
	StarMarkAt     string   `json:"starMarkAt"`
	GivenUpAt      string   `json:"givenUpAt"`
	RemindAt       string   `json:"remindAt"`
	RemindRepeat   string   `json:"remindRepeat"`
	RemindTime     string   `json:"remindTime"`
	RemindWeekdays []uint8  `json:"remindWeekdays"`
	SortId         uint16   `json:"sortId"`
}

// CreateTaskReq 创建任务请求
type CreateTaskReq struct {
	ParentTaskId   string   `json:"parentTaskId"`
	Name           string   `json:"name" binding:"required"`
	Description    string   `json:"description"`
	State          string   `json:"state" binding:"required"`
	Priority       string   `json:"priority" binding:"required"`
	StartAt        string   `json:"startAt"`
	EndAt          string   `json:"endAt"`
	ProjectId      string   `json:"projectId"`
	Tags           []string `json:"tags"`
	RemindAt       string   `json:"remindAt"`
	RemindRepeat   string   `json:"remindRepeat"`
	RemindTime     string   `json:"remindTime"`
	RemindWeekdays []uint8  `json:"remindWeekdays"`
	// 同步元数据：客户端预置 id/createdAt/updatedAt（可空，桌面端同步用）
	Id        *string `json:"id"`
	CreatedAt *string `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
	// DeletedAt 同步元数据：客户端本地删除时间（可空，推送墓碑时携带，服务端 upsert 写入 deleted_at）
	DeletedAt *string `json:"deletedAt"`
}

// UpdateTaskReq 更新任务请求
type UpdateTaskReq struct {
	ParentTaskId   *string  `json:"parentTaskId"`
	Name           *string  `json:"name"`
	Description    *string  `json:"description"`
	State          *string  `json:"state"`
	Priority       *string  `json:"priority"`
	StartAt        *string  `json:"startAt"`
	EndAt          *string  `json:"endAt"`
	ProjectId      *string  `json:"projectId"`
	Tags           []string `json:"tags"`
	ArchivedAt     *string  `json:"archivedAt"`
	StarMarkAt     *string  `json:"starMarkAt"`
	GivenUpAt      *string  `json:"givenUpAt"`
	RemindAt       *string  `json:"remindAt"`
	RemindRepeat   *string  `json:"remindRepeat"`
	RemindTime     *string  `json:"remindTime"`
	RemindWeekdays []uint8  `json:"remindWeekdays"`
	SortId         *uint16  `json:"sortId"`
	// UpdatedAt 乐观锁时间戳（RFC3339，可空）：早于服务端当前版本时不更新（LWW 防旧数据回滚）
	UpdatedAt *string `json:"updatedAt"`
}

// ListTaskReq 列表任务请求
type ListTaskReq struct {
	ParentTaskId string `form:"parentTaskId"`
	ProjectId    string `form:"projectId"`
	TagId        string `form:"tagId"`
	Name         string `form:"name"`
	Description  string `form:"description"`
	State        string `form:"state"`
	Priority     string `form:"priority"`
	StartAt      string `form:"startAt"`
	EndAt        string `form:"endAt"`
	DeletedAt    string `form:"deletedAt"`
	ArchivedAt   string `form:"archivedAt"`
	StarMarkAt   string `form:"starMarkAt"`
	GivenUpAt    string `form:"givenUpAt"`
	IsDeleted    bool   `form:"isDeleted"`
	IsArchived   bool   `form:"isArchived"`
	IsStarMarked bool   `form:"isStarMarked"`
	IsGivenUp    bool   `form:"isGivenUp"`
	RelativeDate string `form:"relativeDate"`
	UpdatedAt    string `form:"updatedAt"`
	CursorId     string `form:"cursorId"`
	ListReqBase
}

// ListTaskRes 列表任务响应
type ListTaskRes []*GetTaskRes

// --- TaskCheckItem ---

// GetTaskCheckItemRes 获取任务检查项响应
type GetTaskCheckItemRes struct {
	ResBase
	TaskId      string `json:"taskId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDone      bool   `json:"isDone"`
	SortId      uint16 `json:"sortId"`
}

// CreateTaskCheckItemReq 创建任务检查项请求
type CreateTaskCheckItemReq struct {
	TaskId      string `json:"taskId" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	// 同步元数据：客户端预置 id/createdAt/updatedAt（可空，桌面端同步用）
	Id        *string `json:"id"`
	CreatedAt *string `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
}

// CreateTaskCheckItemRes 创建任务检查项响应
type CreateTaskCheckItemRes GetTaskCheckItemRes

// UpdateTaskCheckItemReq 更新任务检查项请求
type UpdateTaskCheckItemReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsDone      *bool   `json:"isDone"`
	SortId      *uint16 `json:"sortId"`
	// UpdatedAt 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
	UpdatedAt *string `json:"updatedAt"`
}

// ListTaskCheckItemRes 列表任务检查项响应
type ListTaskCheckItemRes []*GetTaskCheckItemRes

// ResortTaskCheckItemsReq 排序任务检查项请求
type ResortTaskCheckItemsReq struct {
	OriginalId string `json:"originalId"`
	BoundId    string `json:"boundId"`
	Flag       int    `json:"flag"`
}

// ResortTaskCheckItemsRes 排序任务检查项响应
type ResortTaskCheckItemsRes struct {
	OriginalSortId uint16 `json:"originalSortId"`
	BoundSortId    uint16 `json:"boundSortId"`
}

// BatchUpdateTaskCheckItemReq 批量更新任务检查项请求
type BatchUpdateTaskCheckItemReq struct {
	Events []*struct {
		Id          string  `json:"id" binding:"required"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsDone      *bool   `json:"isDone"`
		SortId      *uint16 `json:"sortId"`
	} `json:"events" binding:"required,min=1"`
}

// BatchUpdateTaskCheckItemRes 批量更新任务检查项响应
type BatchUpdateTaskCheckItemRes struct {
	UpdatedCount int64                  `json:"updatedCount"`
	Events       []*GetTaskCheckItemRes `json:"events"`
}

// --- TaskComment ---

// TaskCommentRes 任务评论响应
type TaskCommentRes struct {
	ResBase
	TaskId      string   `json:"taskId"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments"`
	IsTopUp     bool     `json:"isTopUp"`
	Nickname    string   `json:"nickname"`
	Avatar      string   `json:"avatar"`
}

// CreateTaskCommentReq 创建任务评论请求
type CreateTaskCommentReq struct {
	TaskId  string `json:"taskId" binding:"required"`
	Content string `json:"content" binding:"required"`
	// 同步元数据：客户端预置 id/createdAt/updatedAt（可空，桌面端同步用）
	Id        *string `json:"id"`
	CreatedAt *string `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
}

// CreateTaskCommentRes 创建任务评论响应
type CreateTaskCommentRes TaskCommentRes

// UpdateTaskCommentReq 更新任务评论请求
type UpdateTaskCommentReq struct {
	Content     *string   `json:"content"`
	Attachments *[]string `json:"attachments"`
	IsTopUp     *bool     `json:"isTopUp"`
	// UpdatedAt 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
	UpdatedAt *string `json:"updatedAt"`
}

// ListTaskCommentRes 列表任务评论响应
type ListTaskCommentRes []*TaskCommentRes

// --- TaskReminder ---

// SnoozeTaskReq 稍后提醒请求
type SnoozeTaskReq struct {
	DurationMinutes int `json:"durationMinutes" binding:"required,min=1,max=1440"`
}

// SnoozeTaskRes 稍后提醒响应
type SnoozeTaskRes struct {
	RemindAt string `json:"remindAt"`
}
