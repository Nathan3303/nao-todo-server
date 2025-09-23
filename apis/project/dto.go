package apis

import (
	"naotodoserver/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type ProjectPreferenceResponseDTO struct {
	ViewType        string `json:"viewType"`
	GetTodosOptions string `json:"getTodosOptions"`
	Columns         string `json:"columns"`
}

type ProjectResponseDTO struct {
	ID          string                        `json:"id"`
	UserId      string                        `json:"userId"`
	CreatedAt   time.Time                     `json:"createdAt"`
	UpdatedAt   time.Time                     `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt                `json:"deletedAt"`
	IsDeleted   bool                          `json:"isDeleted"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	ArchivedAt  *time.Time                    `json:"archivedAt"`
	IsArchived  bool                          `json:"isArchived"`
	Preference  *ProjectPreferenceResponseDTO `json:"preference"`
}

func ToProjectResponse(project models.Project) ProjectResponseDTO {
	return ProjectResponseDTO{
		ID:          strconv.FormatInt(project.ID, 10),
		UserId:      strconv.FormatInt(project.UserId, 10),
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
		DeletedAt:   project.DeletedAt,
		IsDeleted:   project.DeletedAt.Valid,
		Name:        project.Name,
		Description: project.Description,
		ArchivedAt:  project.ArchivedAt,
		IsArchived:  project.ArchivedAt != nil,
		Preference: &ProjectPreferenceResponseDTO{
			Columns:         project.Preference.Columns,
			GetTodosOptions: project.Preference.GetOptions,
			ViewType:        project.Preference.ViewType,
		},
	}
}
