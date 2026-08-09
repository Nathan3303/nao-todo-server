// Package dto 定义项目应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// GetProjectRes 获取任务清单出参
type GetProjectRes struct {
	Id          string
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
	Name        string
	Description string
	SortId      uint16
	ArchivedAt  string
	DeactivedAt string
}

// CreateProjectReq 创建任务清单入参
type CreateProjectReq struct {
	Name        string
	Description string
	Id          *string // 同步元数据：客户端预置 id
	CreatedAt   *string // 同步元数据：客户端预置 createdAt
	UpdatedAt   *string // 同步元数据：客户端预置 updatedAt
}

// CreateProjectRes 创建任务清单出参
type CreateProjectRes struct {
	Id          string
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
	Name        string
	Description string
	SortId      uint16
	ArchivedAt  string
	DeactivedAt string
}

// UpdateProjectReq 更新任务清单入参
type UpdateProjectReq struct {
	Name        *string
	Description *string
	SortId      *uint16
	UpdatedAt   *string // 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
}

// BatchUpdateProjectItem 批量更新任务清单项
type BatchUpdateProjectItem struct {
	Id          string
	Name        *string
	Description *string
	SortId      *uint16
}

// BatchUpdateProjectReq 批量更新任务清单入参
type BatchUpdateProjectReq struct {
	Projects []*BatchUpdateProjectItem
}

// BatchUpdateProjectRes 批量更新任务清单出参
type BatchUpdateProjectRes struct {
	UpdatedCount int64
	Projects     []*GetProjectRes
}

// ListProjectRes 获取任务清单列表出参
type ListProjectRes []*GetProjectRes

// GetProjectPreferenceRes 获取任务清单偏好出参
type GetProjectPreferenceRes struct {
	Id          string
	CreatedAt   string
	UpdatedAt   string
	DeletedAt   string
	ProjectId   string
	ViewType    string
	GetOptions  string
	Columns     string
}

// UpdateProjectPreferenceReq 更新任务清单偏好入参
type UpdateProjectPreferenceReq struct {
	ViewType   string
	GetOptions string
	Columns    string
}
