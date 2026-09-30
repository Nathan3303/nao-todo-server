package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/types"
)

// CreatePomodoro 创建常用番茄工作值对象
type CreatePomodoro struct {
	Id          int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   types.NullableTime
	UserId      int64
	Type        entities.PomodoroType
	Name        string
	Description string
	Duration    uint16

	// ArchivedAt 归档时间三态（sync push 专用，T466 / DEF-43）：与 CreateProject.ArchivedAt 同口径。
	// Valid=false（缺省；REST create 与旧客户端推送）⇒ 不写列；
	// Valid=true,IsNull=true（null / ""）⇒ 显式清空写 NULL；Valid=true,IsNull=false ⇒ 写入该时间。
	ArchivedAt types.NullableTime

	// BaseUpdatedAt OCC：客户端回传的服务端 updated_at 快照（零值 = 未提供，回退 LWW）
	BaseUpdatedAt time.Time
}

// Validate 验证创建常用番茄工作值对象是否有效
func (vo *CreatePomodoro) Validate() error {
	if vo.Name == "" {
		return errors.New("名称不能为空")
	}
	if vo.Duration <= 0 {
		return errors.New("专注时长必须大于 0")
	}
	return nil
}

// NewCreatePomodoro 创建常用番茄工作值对象
func NewCreatePomodoro(
	userId int64,
	pomodoroType entities.PomodoroType,
	name string,
	description string,
	duration uint16,
) (*CreatePomodoro, error) {
	vo := &CreatePomodoro{
		UserId:      userId,
		Type:        pomodoroType,
		Name:        name,
		Description: description,
		Duration:    duration,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
