package valueobjects

import (
	"errors"
	"time"
)

type CreatePomodoro struct {
	UserId      int64
	SessionId   string
	Type        uint8
	TaskId      int64
	TaskName    string
	Description string
	StartAt     *time.Time
	EndAt       *time.Time
	Duration    int
	Note        string
}

func (vo *CreatePomodoro) Validate() error {
	if vo.SessionId == "" {
		return errors.New("会话 ID 不能为空")
	}
	if vo.TaskId <= 0 {
		return errors.New("任务 ID 无效")
	}
	if vo.TaskName == "" {
		return errors.New("任务名称不能为空")
	}
	if vo.StartAt == nil {
		return errors.New("开始时间不能为空")
	}
	if vo.EndAt == nil {
		return errors.New("结束时间不能为空")
	}
	if vo.Duration <= 0 {
		return errors.New("专注时长必须大于 0")
	}
	if vo.EndAt.Before(*vo.StartAt) {
		return errors.New("结束时间不能早于开始时间")
	}
	return nil
}

func NewCreatePomodoro(
	userId int64,
	sessionId string,
	pomodoroType uint8,
	taskId int64,
	taskName string,
	description string,
	startAt *time.Time,
	endAt *time.Time,
	duration int,
	note string,
) (*CreatePomodoro, error) {
	vo := &CreatePomodoro{
		UserId:      userId,
		SessionId:   sessionId,
		Type:        pomodoroType,
		TaskId:      taskId,
		TaskName:    taskName,
		Description: description,
		StartAt:     startAt,
		EndAt:       endAt,
		Duration:    duration,
		Note:        note,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
