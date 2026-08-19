// Package dto 定义标签应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// GetTagRes 获取标签出参
type GetTagRes struct {
	Id          string
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
	Name        string
	Description string
	Color       string
	SortId      uint16
}

// CreateTagReq 创建标签入参
type CreateTagReq struct {
	Name        string
	Description string
	Color       string
	Id          *string // 同步元数据：客户端预置 id
	CreatedAt   *string // 同步元数据：客户端预置 createdAt
	UpdatedAt   *string // 同步元数据：客户端预置 updatedAt
	DeletedAt   *string // 同步元数据：客户端预置 deletedAt（推送本地墓碑时携带）
}

// CreateTagRes 创建标签出参
type CreateTagRes struct {
	Id          string
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
	Name        string
	Description string
	Color       string
	SortId      uint16
}

// UpdateTagReq 更新标签入参
type UpdateTagReq struct {
	Name        *string
	Description *string
	Color       *string
	SortId      *uint16
	UpdatedAt   *string // 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
}

// BatchUpdateTagItem 批量更新标签项
type BatchUpdateTagItem struct {
	Id          string
	Name        *string
	Description *string
	Color       *string
	SortId      *uint16
}

// BatchUpdateTagReq 批量更新标签入参
type BatchUpdateTagReq struct {
	Tags []*BatchUpdateTagItem
}

// BatchUpdateTagRes 批量更新标签出参
type BatchUpdateTagRes struct {
	UpdatedCount int64
	Tags         []*GetTagRes
}

// GetTagPreferenceRes 获取标签偏好出参
type GetTagPreferenceRes struct {
	Id        string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
	TagId     string
	ViewType  string
	GetOptions string
	Columns    string
}

// UpdateTagPreferenceReq 更新标签偏好入参
type UpdateTagPreferenceReq struct {
	ViewType   string
	GetOptions string
	Columns    string
}
