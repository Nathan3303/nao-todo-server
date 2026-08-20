package types

// --- SyncPush ---

// SyncDeletion 删除墓碑：按表名 + id 软删
type SyncDeletion struct {
	Table string `json:"table" binding:"required"`
	Id    string `json:"id" binding:"required"`
}

// SyncPushReq 批量推送请求
// 各表条目复用对应资源的 create 请求结构（可携带 id/createdAt/updatedAt 同步元数据）
type SyncPushReq struct {
	Tasks           []CreateTaskReq           `json:"tasks"`
	TaskCheckItems  []CreateTaskCheckItemReq  `json:"taskCheckItems"`
	TaskComments    []CreateTaskCommentReq    `json:"taskComments"`
	Projects        []CreateProjectReq        `json:"projects"`
	Tags            []CreateTagReq            `json:"tags"`
	Pomodoros       []CreatePomodoroReq       `json:"pomodoros"`
	PomodoroRecords []CreatePomodoroRecordReq `json:"pomodoroRecords"`
	Deletions       []SyncDeletion            `json:"deletions"`
}

// SyncResult 单条推送结果
type SyncResult struct {
	Table           string `json:"table"`
	Id              string `json:"id"`
	ServerUpdatedAt string `json:"serverUpdatedAt"`
	// Error 本条推送失败原因（部分成功语义：仅失败条目携带，其余为空）
	Error string `json:"error,omitempty"`
	// Skipped 本条被服务端忽略（如只追加资源不支持删除）
	Skipped bool `json:"skipped,omitempty"`
}

// SyncPushRes 批量推送响应
type SyncPushRes struct {
	Results    []SyncResult `json:"results"`
	ServerTime string       `json:"serverTime"`
}

// --- SyncPull ---

// SyncPullTableReq 单表增量拉取请求（updatedAt 可空 = 首次全量）
type SyncPullTableReq struct {
	UpdatedAt string `json:"updatedAt"`
	CursorId  string `json:"cursorId"` // keyset 游标辅助：(updated_at, id) > (updatedAt, cursorId)
	Limit     int    `json:"limit"`
}

// SyncPullReq 批量增量拉取请求
type SyncPullReq struct {
	Tasks           *SyncPullTableReq `json:"tasks"`
	TaskCheckItems  *SyncPullTableReq `json:"taskCheckItems"`
	TaskComments    *SyncPullTableReq `json:"taskComments"`
	Projects        *SyncPullTableReq `json:"projects"`
	Tags            *SyncPullTableReq `json:"tags"`
	Pomodoros       *SyncPullTableReq `json:"pomodoros"`
	PomodoroRecords *SyncPullTableReq `json:"pomodoroRecords"`
}

// SyncPullTableRes 单表增量拉取结果
type SyncPullTableRes struct {
	Items        any    `json:"items"`
	Total        int64  `json:"total"`
	NextCursor   string `json:"nextCursor,omitempty"`   // 本页最后一条 updated_at（RFC3339），空表示无更多
	NextCursorId string `json:"nextCursorId,omitempty"` // 与 NextCursor 组成 keyset 游标
}

// SyncPullRes 批量增量拉取响应
type SyncPullRes struct {
	Data       map[string]SyncPullTableRes `json:"data"`
	ServerTime string                      `json:"serverTime"`
}
