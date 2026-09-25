package types

// --- SyncPush ---

// SyncDeletion 删除墓碑：按表名 + id 软删
type SyncDeletion struct {
	Table string `json:"table" binding:"required"`
	Id    string `json:"id" binding:"required"`
}

// SyncPushReq 批量推送请求
// 各表条目为 sync 专用类型：嵌入对应资源的 create 请求结构（同形）+ additive `baseUpdatedAt`（OCC）。
// ⛔ 不直接复用共享 CreateXxxReq —— 避免把 OCC 字段污染 create REST 的 JSON 契约（C-44）。
type SyncPushReq struct {
	Tasks           []SyncTaskPushItem           `json:"tasks"`
	TaskCheckItems  []SyncCheckItemPushItem      `json:"taskCheckItems"`
	TaskComments    []SyncCommentPushItem        `json:"taskComments"`
	Projects        []SyncProjectPushItem        `json:"projects"`
	Tags            []SyncTagPushItem            `json:"tags"`
	Pomodoros       []SyncPomodoroPushItem       `json:"pomodoros"`
	PomodoroRecords []SyncPomodoroRecordPushItem `json:"pomodoroRecords"`
	Deletions       []SyncDeletion               `json:"deletions"`
}

// 以下 sync 专用条目类型：JSON 层与共享 CreateXxxReq 同形 + 可选 baseUpdatedAt（纯追加）。
// BaseUpdatedAt 为客户端持有的服务端版本快照（RFC3339Milli）；缺失/空 ⇒ 回退现行 LWW（向后兼容）。

type SyncTaskPushItem struct {
	CreateTaskReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
	// 可空时间三态遮蔽位（T319）：以下六个字段与内嵌 CreateTaskReq 的 `*string` 版 JSON 名
	// 完全一致（客户端零改动），但按 encoding/json「浅层优先」规则遮蔽内嵌字段，从而把
	// `null` 与「字段缺省」区分开。语义：absent ⇒ 不写列；null / "" ⇒ 显式清空；值 ⇒ 设值。
	StartAt    NullableString `json:"startAt"`
	EndAt      NullableString `json:"endAt"`
	ArchivedAt NullableString `json:"archivedAt"`
	StarMarkAt NullableString `json:"starMarkAt"`
	GivenUpAt  NullableString `json:"givenUpAt"`
	RemindAt   NullableString `json:"remindAt"`
}

type SyncCheckItemPushItem struct {
	CreateTaskCheckItemReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
}

type SyncCommentPushItem struct {
	CreateTaskCommentReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
}

type SyncProjectPushItem struct {
	CreateProjectReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
}

type SyncTagPushItem struct {
	CreateTagReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
}

type SyncPomodoroPushItem struct {
	CreatePomodoroReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
}

type SyncPomodoroRecordPushItem struct {
	CreatePomodoroRecordReq
	BaseUpdatedAt string `json:"baseUpdatedAt,omitempty"`
}

// SyncResult 单条推送结果
// Outcome 语义（additive，2026-09-24 T143）：与 domain/types.DecideUpsert 判定同源，
// 客户端据此区分「被服务端现有版本覆盖（noop）」与「实际写入（applied）」。
type SyncResult struct {
	Table           string `json:"table"`
	Id              string `json:"id"`
	ServerUpdatedAt string `json:"serverUpdatedAt"`
	// Outcome 本条推送的服务端判定：
	//   applied  = 服务端已写入（新建 / 墓碑复活 / 覆盖）
	//   noop     = 服务端判定请求更旧，未写入，返回库中当前版本（被服务端现有版本覆盖）
	//   stale    = OCC base 不匹配，未写入，返回库中当前版本（additive，2026-09-24 T163；与 conflict 语义区分）
	//   conflict = create 语义 ID 碰撞（同时 error 非空）
	//   skipped  = 服务端忽略该条（如只追加资源不支持删除）
	//   error    = 处理失败（同时 error 非空）
	Outcome string `json:"outcome,omitempty"`
	// Error 本条推送失败原因（部分成功语义：仅失败条目携带，其余为空）
	Error string `json:"error,omitempty"`
	// Skipped 本条被服务端忽略（如只追加资源不支持删除）
	Skipped bool `json:"skipped,omitempty"`
}

// 单条推送结果语义常量（SyncResult.Outcome）
const (
	SyncOutcomeApplied  = "applied"
	SyncOutcomeNoop     = "noop"
	SyncOutcomeStale    = "stale"
	SyncOutcomeConflict = "conflict"
	SyncOutcomeSkipped  = "skipped"
	SyncOutcomeError    = "error"
)

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
