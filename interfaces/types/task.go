package types

type TaskRes struct {
	Id           string   `json:"id"`
	ProjectId    string   `json:"projectId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	State        string   `json:"state"`
	Priority     string   `json:"priority"`
	StartAt      string   `json:"startAt"`
	EndAt        string   `json:"endAt"`
	Tags         []string `json:"tags"`
	UpdatedAt    string   `json:"updatedAt"`
	CreatedAt    string   `json:"createdAt"`
	DeletedAt    string   `json:"deletedAt"`
	IsDeleted    bool     `json:"isDeleted"`
	ArchivedAt   string   `json:"archivedAt"`
	IsArchived   bool     `json:"isArchived"`
	StarMarkAt   string   `json:"starMarkAt"`
	IsStarMarked bool     `json:"isStarMarked"`
	GivenUpAt    string   `json:"givenUpAt"`
	IsGivenUp    bool     `json:"isGivenUp"`
}

type CreateTaskReq struct {
	ProjectId   string   `json:"projectId"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	State       string   `json:"state" binding:"required"`
	Priority    string   `json:"priority" binding:"required"`
	StartAt     string   `json:"startAt"`
	EndAt       string   `json:"endAt" binding:"required"`
	Tags        []string `json:"tags"`
}

type UpdateTaskReq struct {
	ProjectId    string   `json:"projectId"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	State        string   `json:"state"`
	Priority     string   `json:"priority"`
	StartAt      string   `json:"startAt"`
	EndAt        string   `json:"endAt"`
	Tags         []string `json:"tags"`
	IsStarMarked bool     `json:"isStarMarked"`
}

type UpdateTaskRes struct {
	TaskId string `json:"taskId"`
}

type DeleteTaskRes struct {
	TaskId string `json:"taskId"`
}

type RestoreTaskRes struct {
	TaskId string `json:"taskId"`
}

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
	IsArchived   bool   `json:"isArchived"`
	IsStarMarked bool   `json:"isStarMarked"`
	IsGivenUp    bool   `json:"isGivenUp"`
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	RelativeDate string `json:"relativeDate"`
	Sort         string `json:"sort"`
}

type ListTaskRes []*TaskRes
