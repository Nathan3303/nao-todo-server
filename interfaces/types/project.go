package types

// --- Project ---

// CreateProjectReq 创建项目请求
type CreateProjectReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateProjectRes 创建项目响应
type CreateProjectRes struct {
	ResBase
	Name        string `json:"name"`
	Description string `json:"description"`
	SortId      uint16 `json:"sortId"`
	ArchivedAt  string `json:"archivedAt"`
	DeactivedAt string `json:"deactivedAt"`
}

// GetProjectRes 获取项目响应
type GetProjectRes CreateProjectRes

// UpdateProjectReq 更新项目请求
type UpdateProjectReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	SortId      *uint16 `json:"sortId"`
}

// BatchUpdateProjectReq 批量更新项目请求
type BatchUpdateProjectReq struct {
	Projects []*struct {
		Id          string  `json:"id" binding:"required"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		SortId      *uint16 `json:"sortId"`
	} `json:"projects" binding:"required,min=1"`
}

// BatchUpdateProjectRes 批量更新项目响应
type BatchUpdateProjectRes struct {
	UpdatedCount int64            `json:"updatedCount"`
	Projects     []*GetProjectRes `json:"projects"`
}

// ListProjectRes 获取项目列表响应
type ListProjectRes []*GetProjectRes

// --- Project Preference ---

// GetProjectPreferenceRes 获取项目偏好响应
type GetProjectPreferenceRes struct {
	ResBase
	ProjectId  string `json:"projectId"`
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}

// UpdateProjectPreferenceReq 更新项目偏好请求
type UpdateProjectPreferenceReq struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}
