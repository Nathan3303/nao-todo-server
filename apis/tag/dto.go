package apis

import (
	"naotodoserver/models"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type TagPreferenceResponseDTO struct {
	ViewType        string `json:"viewType"`
	GetTodosOptions string `json:"getTodosOptions"`
	Columns         string `json:"columns"`
}

type TagResponseDTO struct {
	ID          string                    `json:"id"`
	UserId      string                    `json:"userId"`
	CreatedAt   time.Time                 `json:"createdAt"`
	UpdatedAt   time.Time                 `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt            `json:"deletedAt"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	ArchivedAt  *time.Time                `json:"archivedAt"`
	Color       string                    `json:"color"`
	IsArchived  bool                      `json:"isArchived"`
	Preference  *TagPreferenceResponseDTO `json:"preference"`
}

func ToTagResponse(tag models.Tag) TagResponseDTO {
	return TagResponseDTO{
		ID:          strconv.FormatInt(tag.ID, 10),
		UserId:      strconv.FormatInt(tag.UserId, 10),
		CreatedAt:   tag.CreatedAt,
		UpdatedAt:   tag.UpdatedAt,
		DeletedAt:   tag.DeletedAt,
		Name:        tag.Name,
		Description: tag.Description,
		ArchivedAt:  tag.ArchivedAt,
		IsArchived:  tag.ArchivedAt != nil,
		Color:       tag.Color,
		Preference: &TagPreferenceResponseDTO{
			Columns:         tag.Preference.Columns,
			GetTodosOptions: tag.Preference.GetOptions,
			ViewType:        tag.Preference.ViewType,
		},
	}
}
