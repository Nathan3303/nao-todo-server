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
	ProjectId    string `json:"projectId" form:"projectId"`
	TagId        string `json:"tagId" form:"tagId"`
	Name         string `json:"name" form:"name"`
	Description  string `json:"description" form:"description"`
	State        string `json:"state" form:"state"`
	Priority     string `json:"priority" form:"priority"`
	StartAt      string `json:"startAt" form:"startAt"`
	EndAt        string `json:"endAt" form:"endAt"`
	DeletedAt    string `json:"deletedAt" form:"deletedAt"`
	ArchivedAt   string `json:"archivedAt" form:"archivedAt"`
	StarMarkAt   string `json:"starMarkAt" form:"starMarkAt"`
	GivenUpAt    string `json:"givenUpAt" form:"givenUpAt"`
	IsDeleted    bool   `json:"isDeleted" form:"isDeleted"`
	IsArchived   bool   `json:"isArchived" form:"isArchived"`
	IsStarMarked bool   `json:"isStarMarked" form:"isStarMarked"`
	IsGivenUp    bool   `json:"isGivenUp" form:"isGivenUp"`
	Page         int    `json:"page" form:"page"`
	Limit        int    `json:"limit" form:"limit"`
	RelativeDate string `json:"relativeDate" form:"relativeDate"`
	Sort         string `json:"sort" form:"sort"`
}

type ListTaskRes []*TaskRes
