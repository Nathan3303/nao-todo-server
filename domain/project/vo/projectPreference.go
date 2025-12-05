package vo

import (
	"strings"
	"time"
)

type ProjectPreference struct {
	Id         int64      `json:"id"`
	CreatedAt  *time.Time `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt"`
	UserId     int64      `json:"userId"`
	ProjectId  int64      `json:"projectId"`
	ViewType   string     `json:"viewType"`
	GetOptions string     `json:"getTodosOptions"`
	Columns    string     `json:"columns"`
}

func MakeDefaultProjectPreference(userId int64) *ProjectPreference {
	return &ProjectPreference{
		ViewType:   "table",
		GetOptions: "{\"limit\": 20}",
		Columns: strings.Join([]string{
			"{",
			"\"createdAt\":false,",
			"\"updatedAt\":false,",
			"\"description\":true,",
			"\"state\":true,",
			"\"priority\":true,",
			"\"tags\":true,",
			"\"startAt\":false,",
			"\"endAt\":true",
			"}",
		}, ""),
	}
}
