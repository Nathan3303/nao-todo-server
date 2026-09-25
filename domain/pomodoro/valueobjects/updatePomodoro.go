package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/types"
)

// UpdatePomodoro 更新常用番茄工作值对象
// 指针字段表示 PATCH 语义：nil 时不更新该字段
type UpdatePomodoro struct {
	Type        *entities.PomodoroType
	Name        *string
	Description *string
	Duration    *uint16
	ArchivedAt  types.NullableTime
	// UpdatedAt 乐观锁时间戳（零值=未提供）：早于服务端当前版本时不更新（LWW）
	UpdatedAt time.Time
}

// Validate 验证更新常用番茄工作值对象是否有效
func (vo *UpdatePomodoro) Validate() error {
	if vo.Name != nil && *vo.Name == "" {
		return errors.New("名称不能为空")
	}
	if vo.Duration != nil && *vo.Duration <= 0 {
		return errors.New("专注时长必须大于 0")
	}
	return nil
}

// NewUpdatePomodoro 创建更新常用番茄工作值对象
func NewUpdatePomodoro(
	pomodoroType *entities.PomodoroType,
	name *string,
	description *string,
	duration *uint16,
	archivedAt string,
) (*UpdatePomodoro, error) {
	vo := &UpdatePomodoro{
		Type:        pomodoroType,
		Name:        name,
		Description: description,
		Duration:    duration,
		ArchivedAt:  types.NewNullableTimeByTimeStr(archivedAt),
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
