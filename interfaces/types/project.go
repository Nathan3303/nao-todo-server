package types

import "time"

type CreateProjectReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateProjectRes struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ArchivedAt  string    `json:"archivedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	DeactivedAt string    `json:"deactivedAt"`
	SortId      uint16    `json:"sortId"`
}

type GetProjectRes struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ArchivedAt  string    `json:"archivedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	DeactivedAt string    `json:"deactivedAt"`
	SortId      uint16    `json:"sortId"`
}

type UpdateProjectReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	SortId      *uint16 `json:"sortId"`
}

type BatchUpdateProjectReq struct {
	Projects []*struct {
		Id          string  `json:"id" binding:"required"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		SortId      *uint16 `json:"sortId"`
	} `json:"projects" binding:"required,min=1"`
}

type BatchUpdateProjectRes struct {
	UpdatedCount int64            `json:"updatedCount"`
	Projects     []*GetProjectRes `json:"projects"`
}

type ListProjectRes []*GetProjectRes

type GetProjectPreferenceRes struct {
	Id         string    `json:"id"`
	ProjectId  string    `json:"projectId"`
	ViewType   string    `json:"viewType"`
	GetOptions string    `json:"getTasksOptions"`
	Columns    string    `json:"columns"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type UpdateProjectPreferenceReq struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}
