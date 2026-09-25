package types

// --- Tag ---

// CreateTagReq 创建标签请求
type CreateTagReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color" binding:"required"`
	// 同步元数据：客户端预置 id/createdAt/updatedAt（可空，桌面端同步用）
	Id        *string `json:"id"`
	CreatedAt *string `json:"createdAt"`
	UpdatedAt *string `json:"updatedAt"`
	// DeletedAt 同步元数据：客户端本地删除时间（可空，推送墓碑时携带，服务端 upsert 写入 deleted_at）
	DeletedAt *string `json:"deletedAt"`
}

// CreateTagRes 创建标签响应
type CreateTagRes struct {
	ResBase
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	SortId      uint16 `json:"sortId"`
}

// GetTagRes 获取标签响应
type GetTagRes CreateTagRes

// UpdateTagReq 更新标签请求
type UpdateTagReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
	SortId      *uint16 `json:"sortId"`
	// UpdatedAt 乐观锁时间戳（RFC3339，可空）：早于服务端版本时不更新
	UpdatedAt *string `json:"updatedAt"`
}

// BatchUpdateTagReq 批量更新标签请求
type BatchUpdateTagReq struct {
	Tags []*struct {
		Id          string  `json:"id" binding:"required"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		SortId      *uint16 `json:"sortId"`
	} `json:"tags" binding:"required,min=1"`
}

// BatchUpdateTagRes 批量更新标签响应
type BatchUpdateTagRes struct {
	UpdatedCount int64        `json:"updatedCount"`
	Tags         []*GetTagRes `json:"tags"`
}

// ListTagRes 获取标签列表响应
type ListTagRes []*GetTagRes

// --- Tag Preference ---

// GetTagPreferenceRes 获取标签偏好响应
type GetTagPreferenceRes struct {
	ResBase
	TagId      string `json:"tagId"`
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}

// UpdateTagPreferenceReq 更新标签偏好请求
type UpdateTagPreferenceReq struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTasksOptions"`
	Columns    string `json:"columns"`
}
