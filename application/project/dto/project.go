// Package dto 定义项目应用层的入参与出参结构体
// @description 应用层自有 DTO，不含任何 HTTP 关注点（如 json / form tag）
package dto

// GetProjectRes 获取任务清单出参
// 注意：与 CreateProjectRes 保持字段一致（appImpl 有直接类型转换）
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
	TaskCount   uint // 项目任务数量（服务端 owned）
}

// CreateProjectReq 创建任务清单入参
type CreateProjectReq struct {
	Name        string
	Description string
	Id          *string // 同步元数据：客户端预置 id
	CreatedAt   *string // 同步元数据：客户端预置 createdAt
	UpdatedAt   *string // 同步元数据：客户端预置 updatedAt
	DeletedAt   *string // 同步元数据：客户端预置 deletedAt（推送本地墓碑时携带）
	// ArchivedAt 同步专用（T322 / DEF-42）：nil=缺省不写列；""=显式清空（写 NULL）；时间串=写入
	// json:"-" 关闭 JSON 注入面：该字段仅由 sync 控制器在 Go 侧显式赋值，create REST 不可经请求体注入
	ArchivedAt *string `json:"-"`
	// BaseUpdatedAt OCC：客户端回传的服务端 updated_at 快照（nil/空 = 未提供，回退 LWW）
	// json:"-" 关闭 JSON 注入面：该字段仅由 sync 控制器在 Go 侧显式赋值，create REST 不可经请求体注入
	BaseUpdatedAt *string `json:"-"`
}

// CreateProjectRes 创建任务清单出参
// 注意：与 GetProjectRes 保持字段一致（appImpl 有直接类型转换）
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
	TaskCount   uint // 项目任务数量（服务端 owned）
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
	Id         string
	CreatedAt  string
	UpdatedAt  string
	DeletedAt  string
	ProjectId  string
	ViewType   string
	GetOptions string
	Columns    string
}

// UpdateProjectPreferenceReq 更新任务清单偏好入参
type UpdateProjectPreferenceReq struct {
	ViewType   string
	GetOptions string
	Columns    string
}
