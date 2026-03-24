package vo

import "strings"

type TagPreference struct {
	Id         int64  `json:"id"`
	UserId     int64  `json:"userId"`
	TagId      int64  `json:"tagId"`
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTodosOptions"`
	Columns    string `json:"columns"`
}

func MakeDefaultTagPreference() *TagPreference {
	return &TagPreference{
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
