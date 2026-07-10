package types

// --- ResponseData ---

// ResponseData 响应数据结构体
type ResponseData struct {
	Code       int         `json:"code"`
	Message    string      `json:"message,omitempty"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Status     int         `json:"status,omitempty"`
	Error      string      `json:"error,omitempty"` // 错误用
}

// Pagination 分页数据结构体
type Pagination struct {
	Total   int64 `json:"total"`   // 总记录数
	Page    int   `json:"page"`    // 当前页数
	Limit   int   `json:"limit"`   // 每页条数
	MaxPage int   `json:"maxPage"` // 最大页数
}

// NewResolvedResponseData 成功响应数据结构体工厂函数
func NewResolvedResponseData(code int, message string, data any) *ResponseData {
	var rd ResponseData
	rd.Code = code
	rd.Message = message
	rd.Data = data
	return &rd
}

// NewRejectedResponseData 失败响应数据结构体工厂函数
func NewRejectedResponseData(code int, errorMsg string, data any) *ResponseData {
	var rd ResponseData
	rd.Code = code
	rd.Error = errorMsg
	rd.Data = data
	return &rd
}

// --- ResBase ---

// ResBase 基本响应数据结构体
type ResBase struct {
	Id        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	DeletedAt string `json:"deletedAt"`
}

// --- ListReqBase ---

// ListReqBase 列表基础请求数据结构体
type ListReqBase struct {
	Page  int    `json:"page" form:"page"`
	Limit int    `json:"limit" form:"limit"`
	Sort  string `json:"sort" form:"sort"`
}
