package types

type ResponseData struct {
	Code       int         `json:"code"`
	Message    string      `json:"message,omitempty"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Status     int         `json:"status,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type Pagination struct {
	Total   int64 `json:"total"`   // 总记录数
	Page    int   `json:"page"`    // 当前页数
	Limit   int   `json:"limit"`   // 每页条数
	MaxPage int   `json:"maxPage"` // 最大页数
}
