package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/pomodoro/entities"
)

// CreatePomodoro 创建常用番茄工作值对象
type CreatePomodoro struct {
	Id          int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UserId      int64
	Type        entities.PomodoroType
	Name        string
	Description string
	Duration    uint16
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
