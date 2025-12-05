package types

import "time"

type CreateProjectReq struct {
	Name        string `json:"name" form:"name" binding:"required"`
	Description string `json:"description" form:"description"`
}

type CreateProjectRes struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	Preference  any        `json:"preference"`
}

type GetProjectReq struct {
	ProjectId string
}

type GetProjectRes struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ArchivedAt  *time.Time `json:"archivedAt"`
	Preference  any        `json:"preference"`
}

type UpdateProjectReq struct {
	ProjectId   string
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
}

type UpdateProjectRes struct {
	ProjectId string `json:"projectId"`
}

type DeleteProjectReq struct {
	ProjectId string
}

type DeleteProjectRes struct {
	ProjectId string `json:"projectId"`
}

type RestoreProjectReq struct {
	ProjectId string
}

type RestoreProjectRes struct {
	ProjectId string `json:"projectId"`
}

type ArchiveProjectReq struct {
	ProjectId string
}

type ArchiveProjectRes struct {
	ProjectId string `json:"projectId"`
}

type UnarchiveProjectReq struct {
	ProjectId string
}

type UnarchiveProjectRes struct {
	ProjectId string `json:"projectId"`
}

type HardDeleteProjectReq struct {
	ProjectId string
}

type HardDeleteProjectRes struct {
	ProjectId string `json:"projectId"`
}

type UpdateProjectPreferenceReq struct {
	ProjectId  string
	Preference any `json:"preference" form:"preference"`
}

type UpdateProjectPreferenceRes struct {
	ProjectId string `json:"projectId"`
}

type ListProjectRes []*GetProjectRes

type ProjectPreferenceRes struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getOptions"`
	Columns    string `json:"columns"`
}
