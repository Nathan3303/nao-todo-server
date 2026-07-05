package valueobjects

// QueryPomodoroRecord 番茄工作记录查询值对象
type QueryPomodoroRecord struct {
	UserId    int64
	SessionId string
	StartTime string
	EndTime   string
	TaskId    int64
	TaskName  string
	Type      uint8
	Sort      string
	Page      int
	Limit     int
}

// NewQueryPomodoroRecord 创建 QueryPomodoroRecord 值对象
func NewQueryPomodoroRecord(
	userId int64,
	sessionId string,
	startTime string,
	endTime string,
	taskId int64,
	taskName string,
	pomodoroType uint8,
	sort string,
	page int,
	limit int,
) *QueryPomodoroRecord {
	return &QueryPomodoroRecord{
		UserId:    userId,
		SessionId: sessionId,
		StartTime: startTime,
		EndTime:   endTime,
		TaskId:    taskId,
		TaskName:  taskName,
		Type:      pomodoroType,
		Sort:      sort,
		Page:      page,
		Limit:     limit,
	}
}
