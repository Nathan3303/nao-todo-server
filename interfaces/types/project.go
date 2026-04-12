package types

import "time"

type CreateProjectReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CreateProjectRes struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type GetProjectRes struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type UpdateProjectReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
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
