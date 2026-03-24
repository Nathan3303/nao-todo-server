package types

type GetTagRes struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Color       string            `json:"color"`
	Preference  *TagPreferenceRes `json:"preference"`
}

type TagPreferenceRes struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}

type CreateTagReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type CreateTagRes struct {
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Color       string            `json:"color"`
	Preference  *TagPreferenceRes `json:"preference"`
}

type UpdateTagReq struct {
	TagId       string
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateTagRes struct {
	TagId string `json:"tagId"`
}

type DeleteTagRes struct {
	TagId string `json:"tagId"`
}

type ListTagRes []*GetTagRes

type TagPreferenceReq struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}

type UpdateTagPreferenceReq struct {
	Preference *TagPreferenceReq `json:"preference"`
}

type UpdateTagPreferenceRes struct {
	TagId string `json:"tagId"`
}
