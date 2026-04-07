package types

import "time"

// GetTaskRes 获取任务响应
type GetTaskRes struct {
	Id           string    `json:"id"`
	ParentTaskId string    `json:"parentTaskId"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	Priority     string    `json:"priority"`
	StartAt      string    `json:"startAt"`
	EndAt        string    `json:"endAt"`
	Tags         []string  `json:"tags"`
	ProjectId    string    `json:"projectId"`
	ArchivedAt   string    `json:"archivedAt"`
	StarMarkAt   string    `json:"starMarkAt"`
	GivenUpAt    string    `json:"givenUpAt"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    time.Time `json:"deletedAt"`
}

// CreateTaskReq 创建任务请求
type CreateTaskReq struct {
	ParentTaskId string   `json:"parentTaskId"`
	Name         string   `json:"name" binding:"required"`
	Description  string   `json:"description"`
	State        string   `json:"state" binding:"required"`
	Priority     string   `json:"priority" binding:"required"`
	StartAt      string   `json:"startAt"`
	EndAt        string   `json:"endAt"`
	ProjectId    string   `json:"projectId"`
	Tags         []string `json:"tags"`
}

// UpdateTaskReq 更新任务请求
type UpdateTaskReq struct {
	ParentTaskId string   `json:"parentTaskId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	State        string   `json:"state"`
	Priority     string   `json:"priority"`
	StartAt      string   `json:"startAt"`
	EndAt        string   `json:"endAt"`
	ProjectId    string   `json:"projectId"`
	Tags         []string `json:"tags"`
	ArchivedAt   string   `json:"archivedAt"`
	StarMarkAt   string   `json:"starMarkAt"`
	GivenUpAt    string   `json:"givenUpAt"`
}

// ListTaskReq 列表任务请求
type ListTaskReq struct {
	ProjectId    string `json:"projectId"`
	TagId        string `json:"tagId"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	State        string `json:"state"`
	Priority     string `json:"priority"`
	StartAt      string `json:"startAt"`
	EndAt        string `json:"endAt"`
	DeletedAt    string `json:"deletedAt"`
	ArchivedAt   string `json:"archivedAt"`
	StarMarkAt   string `json:"starMarkAt"`
	GivenUpAt    string `json:"givenUpAt"`
	IsDeleted    bool   `json:"isDeleted"`
	IsArchived   bool   `json:"isArchived"`
	IsStarMarked bool   `json:"isStarMarked"`
	IsGivenUp    bool   `json:"isGivenUp"`
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	RelativeDate string `json:"relativeDate"`
	Sort         string `json:"sort"`
}

// ListTaskRes 列表任务响应
type ListTaskRes []*GetTaskRes
