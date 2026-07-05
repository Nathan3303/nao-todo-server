package valueobjects

// QueryPomodoro 常用番茄工作查询值对象
type QueryPomodoro struct {
	UserId     int64
	Type       uint8
	Name       string
	IsArchived bool
	Sort       string
	Page       int
	Limit      int
}

// NewQueryPomodoro 创建 QueryPomodoro 值对象
func NewQueryPomodoro(
	userId int64,
	pomodoroType uint8,
	name string,
	isArchived bool,
	sort string,
	page int,
	limit int,
) *QueryPomodoro {
	return &QueryPomodoro{
		UserId:     userId,
		Type:       pomodoroType,
		Name:       name,
		IsArchived: isArchived,
		Sort:       sort,
		Page:       page,
		Limit:      limit,
	}
}
