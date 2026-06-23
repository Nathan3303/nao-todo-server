package types

// GetTaskRes 获取任务响应
type GetTaskRes struct {
	Id             string   `json:"id"`
	UserId         string   `json:"userId"`
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
	RemindWeekdays []int    `json:"remindWeekdays"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
	DeletedAt      string   `json:"deletedAt"`
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
	RemindWeekdays []int    `json:"remindWeekdays"`
}

// UpdateTaskReq 更新任务请求
type UpdateTaskReq struct {
	ParentTaskId   *string        `json:"parentTaskId"`
	Name           *string        `json:"name"`
	Description    *string        `json:"description"`
	State          *string        `json:"state"`
	Priority       *string        `json:"priority"`
	StartAt        NullableString `json:"startAt"`
	EndAt          NullableString `json:"endAt"`
	ProjectId      *string        `json:"projectId"`
	Tags           []string       `json:"tags"`
	ArchivedAt     NullableString `json:"archivedAt"`
	StarMarkAt     NullableString `json:"starMarkAt"`
	GivenUpAt      NullableString `json:"givenUpAt"`
	RemindAt       NullableString `json:"remindAt"`
	RemindRepeat   *string        `json:"remindRepeat"`
	RemindTime     *string        `json:"remindTime"`
	RemindWeekdays []int          `json:"remindWeekdays"`
}

// ListTaskReq 列表任务请求
type ListTaskReq struct {
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
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
	RelativeDate string `form:"relativeDate"`
	Sort         string `form:"sort"`
}

// ListTaskRes 列表任务响应
type ListTaskRes []*GetTaskRes
